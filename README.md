# Go backend for Cron

- Check [main repo](https://github.com/ushiradineth/cron) for details about the project.

## Running the Project Locally

### Clone the repository

- `git clone https://github.com/ushiradineth/cron-be`

### Environment variables

- Check the `.env.example` file for the required environment variables.
- Use `cp .env.example .env` to create the `.env` file.

### Start the MySQL Database

- Run `docker-compose -f docker-compose.yml up -d` to start the MySQL Database and Adminer.
- Wait for a moment for the database to initialize.

### Connect to the database

- You can use Adminer, a web-based administration tool included in the setup, to manage your database. Access Adminer at [localhost:9090](http://localhost:9090).
- In Adminer, use the following credentials:
  - Server: mysql:3306
  - Username: cron
  - Password: password
  - Database: cron

### Run the Go Server

- Run `go run .` to start the Cron Go HTTP Server.

## Build Docker Image

### Create Environment variables

- Check the `.env.example` file for the required environment variables.
- Use `cp .env.example .env.production` to create the `.env.production` file.

### Build the image

- Run `docker build -t cron-be:prod -f Dockerfile .` to build the image.

### Run the image using Docker Compose

- Uncomment the `cron-be` service in `docker-compose.yml`.
- Run `docker-compose -f docker-compose.yml up -d` to start the MySQL Database, Adminer, and the Cron Go HTTP Server.
