const TICKET_API = "/api/tickets";
const USER_API = "/api/users";


/*
 * Load users
 */

async function loadUsers() {

    const response = await fetch(
        USER_API
    );

    const users = await response.json();

    const select =
        document.getElementById("user");


    select.innerHTML = "";


    users.forEach(user => {

        const option =
            document.createElement("option");


        option.value = user.id;


        option.textContent =
            `${user.name} (${user.email})`;


        select.appendChild(option);

    });

}


/*
 * Load tickets
 */

async function loadTickets() {

    const response = await fetch(
        TICKET_API
    );

    const tickets =
        await response.json();


    const container =
        document.getElementById("tickets");


    container.innerHTML = "";


    tickets.forEach(ticket => {

        const div =
            document.createElement("div");


        div.className = "ticket";


        div.innerHTML = `

            <h3>
                ${ticket.subject}
            </h3>

            <p>
                ${ticket.description}
            </p>

            <p>
                User ID:
                ${ticket.user_id}
            </p>

            <p>
                Status:
                ${ticket.status}
            </p>

        `;


        container.appendChild(div);

    });

}


/*
 * Create ticket
 */

document
    .getElementById("ticket-form")
    .addEventListener(
        "submit",
        async event => {

            event.preventDefault();


            const user_id =
                document.getElementById(
                    "user"
                ).value;


            const subject =
                document.getElementById(
                    "subject"
                ).value;


            const description =
                document.getElementById(
                    "description"
                ).value;


            const response =
                await fetch(
                    TICKET_API,
                    {

                        method: "POST",

                        headers: {
                            "Content-Type":
                                "application/json"
                        },

                        body: JSON.stringify({

                            user_id:
                                Number(user_id),

                            subject,

                            description

                        })

                    }
                );


            const data =
                await response.json();


            if (!response.ok) {

                alert(
                    data.error ||
                    "Could not create ticket"
                );

                return;

            }


            alert(
                "Ticket created successfully"
            );


            document
                .getElementById("ticket-form")
                .reset();


            loadTickets();

        }
    );


/*
 * Initial load
 */

loadUsers();

loadTickets();