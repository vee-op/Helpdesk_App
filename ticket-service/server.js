const express = require("express");
const { Pool } = require("pg");
const { createClient } = require("redis");

const app = express();

app.use(express.json());

const PORT = process.env.PORT || 5000;

const USER_SERVICE_URL =
    process.env.USER_SERVICE_URL;


/*
 * PostgreSQL
 */

const pool = new Pool({

    host: process.env.DB_HOST,

    port: process.env.DB_PORT,

    database: process.env.DB_NAME,

    user: process.env.DB_USER,

    password: process.env.DB_PASSWORD
});


/*
 * Redis
 */

const redis = createClient({

    url:
        process.env.REDIS_URL 
});


redis.on("error", error => {

    console.error(
        "Redis error:",
        error
    );

});


/*
 * Initialize
 */

async function initialize() {

    await redis.connect();

    await pool.query(`
        CREATE TABLE IF NOT EXISTS tickets (

            id SERIAL PRIMARY KEY,

            user_id INTEGER NOT NULL,

            subject VARCHAR(200) NOT NULL,

            description TEXT NOT NULL,

            status VARCHAR(50)
                DEFAULT 'open',

            created_at TIMESTAMP
                DEFAULT CURRENT_TIMESTAMP
        )
    `);

    console.log(
        "Ticket database initialized"
    );
}


/*
 * Health check
 */

app.get("/health", async (req, res) => {

    try {

        await pool.query("SELECT 1");

        res.json({

            service: "ticket-service",

            status: "healthy"

        });

    } catch (error) {

        res.status(500).json({

            service: "ticket-service",

            status: "unhealthy"

        });

    }

});


/*
 * Get all tickets
 */

app.get("/tickets", async (req, res) => {

    try {

        const result = await pool.query(`
            SELECT *
            FROM tickets
            ORDER BY created_at DESC
        `);

        res.json(result.rows);

    } catch (error) {

        console.error(error);

        res.status(500).json({

            error:
                "Could not retrieve tickets"

        });

    }

});


/*
 * Create ticket
 */

app.post("/tickets", async (req, res) => {

    const {
        user_id,
        subject,
        description
    } = req.body;


    /*
     * Validate request
     */

    if (
        !user_id ||
        !subject ||
        !description
    ) {

        return res.status(400).json({

            error:
                "user_id, subject and description are required"

        });

    }


    /*
     * Ask User Service
     *
     * This is the important part.
     */

    try {

        const response = await fetch(
            `${USER_SERVICE_URL}/users/${user_id}`
        );


        if (!response.ok) {

            if (response.status === 404) {

                return res.status(400).json({

                    error:
                        "User does not exist"

                });

            }

            return res.status(502).json({

                error:
                    "User service unavailable"

            });

        }


        const user = await response.json();


        /*
         * User exists.
         * Now create the ticket.
         */

        const result = await pool.query(
            `
            INSERT INTO tickets
            (user_id, subject, description)
            VALUES ($1, $2, $3)
            RETURNING *
            `,
            [
                user.id,
                subject,
                description
            ]
        );


        const ticket = result.rows[0];


        /*
         * Publish event to Redis
         */

        await redis.publish(

            "ticket-created",

            JSON.stringify({

                ticket,

                user

            })

        );


        /*
         * Return combined information
         */

        res.status(201).json({

            message:
                "Ticket created successfully",

            ticket,

            user

        });


    } catch (error) {

        console.error(error);

        res.status(500).json({

            error:
                "Could not create ticket"

        });

    }

});


/*
 * Start service
 */

initialize()

    .then(() => {

        app.listen(
            PORT,
            () => {

                console.log(
                    `Ticket service running on port ${PORT}`
                );

            }
        );

    })

    .catch(error => {

        console.error(
            "Failed to initialize service:",
            error
        );

    });