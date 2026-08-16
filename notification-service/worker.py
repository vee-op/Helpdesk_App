import json
import os
import redis


REDIS_URL = os.getenv(
    "REDIS_URL",
    ""
)


def main():

    client = redis.from_url(
        REDIS_URL,
        decode_responses=True
    )

    pubsub = client.pubsub()

    pubsub.subscribe("ticket-created")

    print(
        "Notification service started"
    )

    print(
        "Waiting for ticket-created events..."
    )


    for message in pubsub.listen():

        if message["type"] != "message":
            continue


        data = json.loads(
            message["data"]
        )


        ticket = data["ticket"]

        user = data["user"]


        print("\n==========================")

        print(
            "NEW TICKET"
        )

        print(
            f"Ticket ID: {ticket['id']}"
        )

        print(
            f"User: {user['name']}"
        )

        print(
            f"Email: {user['email']}"
        )

        print(
            f"Subject: {ticket['subject']}"
        )

        print(
            f"Description: {ticket['description']}"
        )

        print(
            "==========================\n"
        )


if __name__ == "__main__":

    main()