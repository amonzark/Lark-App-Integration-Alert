# Lark App

A minimalistic Go application designed to seamlessly integrate Alertmanager with Lark Messaging.

## Prerequisites

Below are the prerequisites for the Katulampa Lark App. The specified versions are used in development, and it is recommended to use the same versions for consistency.

- go (1.23)
- docker (26.1.5)

## Development

To start the application in development mode, execute the command `make dev` using the provided Makefile.
Please do update the `alertmanager.yml` file with the appropriate `group_id` for the intended channel to successfully send the test alert.

## Environment Variables
The following table lists the environment variables related to the Katulampa Lark App:
| Environment Variable | Description                         | Default Value | Required |
|----------------------|-------------------------------------|---------------|----------|
| `LARK_APP_ID`        | The App ID for Lark integration     | `""`          | Yes      |
| `LARK_APP_SECRET`    | The App Secret for Lark integration | `""`          | Yes      |
| `REDIS_URL`          | The URL for the Redis instance      | `""`          | Yes      |
| `REDIS_PASSWORD`     | The password for the Redis instance | `""`          | No       |




