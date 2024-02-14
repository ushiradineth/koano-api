# Go backend for Cron

- Check [main repo](https://github.com/ushiradineth/cron) for details about the project.

## Running the Project

### Clone the repository

- `git clone https://github.com/ushiradineth/cron-be`

### Environment variables

- Follow the `.env.example` for the required environment variable.
- Use `cp .env.example .env` to create the .env file.

### Start the MySQL Database

- Run `docker-compose -f docker_compose.yml up -d` to start the and database container in the background.

### Connect to the database

- You can use Adminer, a web-based administration tool included in the setup, to manage your database. Access Adminer at [localhost:9090](http://localhost:9090).

- In Adminer, use the following credentials:
  - Server: mysql:3306
  - Username: cron
  - Password: password
  - database: cron

### Run the Go Server

- Run `go run .` to start the Cron Go HTTP Server.
