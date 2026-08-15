package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

var db *sql.DB

func main() {

	/*
	 * Database configuration
	 */

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "users")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")


	/*
	 * PostgreSQL connection
	 */

	connectionString :=
		"host=" + dbHost +
			" port=" + dbPort +
			" dbname=" + dbName +
			" user=" + dbUser +
			" password=" + dbPassword +
			" sslmode=disable"


	var err error

	db, err = sql.Open(
		"postgres",
		connectionString,
	)

	if err != nil {
		log.Fatal(err)
	}


	/*
	 * Test database connection
	 */

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")


	/*
	 * Create database table
	 */

	initializeDatabase()


	/*
	 * HTTP routes
	 */

	router := mux.NewRouter()

	router.HandleFunc(
		"/health",
		healthHandler,
	).Methods("GET")

	router.HandleFunc(
		"/users",
		getUsers,
	).Methods("GET")

	router.HandleFunc(
		"/users/{id}",
		getUser,
	).Methods("GET")

	router.HandleFunc(
		"/users",
		createUser,
	).Methods("POST")


	/*
	 * Start server
	 */

	port := getEnv(
		"PORT",
		"5001",
	)

	log.Printf(
		"User service running on port %s",
		port,
	)

	log.Fatal(
		http.ListenAndServe(
			":"+port,
			router,
		),
	)
}


/*
 * Database initialization
 */

func initializeDatabase() {

	query := `
	CREATE TABLE IF NOT EXISTS users (

		id SERIAL PRIMARY KEY,

		name VARCHAR(100) NOT NULL,

		email VARCHAR(150) UNIQUE NOT NULL,

		role VARCHAR(50)
			DEFAULT 'customer'
	)
	`


	_, err := db.Exec(query)

	if err != nil {

		log.Fatal(
			"Failed to create users table:",
			err,
		)
	}


	/*
	 * Add initial users
	 * only when the table is empty.
	 */

	var count int

	err = db.QueryRow(
		"SELECT COUNT(*) FROM users",
	).Scan(&count)

	if err != nil {

		log.Fatal(
			"Failed to count users:",
			err,
		)
	}


	if count == 0 {

		_, err = db.Exec(`
			INSERT INTO users
				(name, email, role)
			VALUES
				(
					'John Admin',
					'john@example.com',
					'admin'
				),
				(
					'Sarah Support',
					'sarah@example.com',
					'support'
				),
				(
					'Mike Support',
					'mike@example.com',
					'support'
				)
		`)

		if err != nil {

			log.Fatal(
				"Failed to insert initial users:",
				err,
			)
		}

		log.Println(
			"Initial users created",
		)

	} else {

		log.Println(
			"Users already exist",
		)
	}
}


/*
 * Health check
 *
 * GET /health
 */

func healthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	err := db.Ping()

	if err != nil {

		w.WriteHeader(
			http.StatusInternalServerError,
		)

		json.NewEncoder(w).Encode(
			map[string]string{
				"service": "user-service",
				"status":  "unhealthy",
			},
		)

		return
	}


	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]string{
			"service": "user-service",
			"status":  "healthy",
		},
	)
}


/*
 * Get all users
 *
 * GET /users
 */

func getUsers(
	w http.ResponseWriter,
	r *http.Request,
) {

	rows, err := db.Query(`
		SELECT
			id,
			name,
			email,
			role
		FROM users
		ORDER BY id
	`)

	if err != nil {

		http.Error(
			w,
			"Could not retrieve users",
			http.StatusInternalServerError,
		)

		return
	}

	defer rows.Close()


	var users []User


	for rows.Next() {

		var user User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Role,
		)

		if err != nil {

			http.Error(
				w,
				"Database error",
				http.StatusInternalServerError,
			)

			return
		}


		users = append(
			users,
			user,
		)
	}


	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(users)
}


/*
 * Get one user
 *
 * GET /users/{id}
 *
 * This is the endpoint that the
 * Ticket Service will call.
 */

func getUser(
	w http.ResponseWriter,
	r *http.Request,
) {

	id := mux.Vars(r)["id"]


	var user User


	err := db.QueryRow(`
		SELECT
			id,
			name,
			email,
			role
		FROM users
		WHERE id = $1
	`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
	)


	/*
	 * User does not exist
	 */

	if err == sql.ErrNoRows {

		http.Error(
			w,
			"User not found",
			http.StatusNotFound,
		)

		return
	}


	/*
	 * Other database error
	 */

	if err != nil {

		http.Error(
			w,
			"Database error",
			http.StatusInternalServerError,
		)

		return
	}


	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(user)
}


/*
 * Create user
 *
 * POST /users
 */

func createUser(
	w http.ResponseWriter,
	r *http.Request,
) {

	var user User


	err := json.NewDecoder(
		r.Body,
	).Decode(&user)

	if err != nil {

		http.Error(
			w,
			"Invalid request",
			http.StatusBadRequest,
		)

		return
	}


	/*
	 * Basic validation
	 */

	if user.Name == "" ||
		user.Email == "" {

		http.Error(
			w,
			"Name and email are required",
			http.StatusBadRequest,
		)

		return
	}


	/*
	 * Default role
	 */

	if user.Role == "" {

		user.Role = "customer"
	}


	/*
	 * Insert user
	 */

	err = db.QueryRow(`
		INSERT INTO users
			(name, email, role)
		VALUES
			($1, $2, $3)
		RETURNING id
	`,
		user.Name,
		user.Email,
		user.Role,
	).Scan(
		&user.ID,
	)


	if err != nil {

		http.Error(
			w,
			"Could not create user",
			http.StatusInternalServerError,
		)

		return
	}


	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(
		http.StatusCreated,
	)

	json.NewEncoder(w).Encode(user)
}


/*
 * Environment variable helper
 */

func getEnv(
	key string,
	fallback string,
) string {

	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}