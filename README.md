# CSE 312 Project

## Overview

This project is a web application built using Go and the Gin web framework. It uses PostgreSQL for database storage, Docker Compose to manage the application services, and Nginx as a reverse proxy.

## Technologies Used

- **Go** – Backend programming language
- **Gin** – Web framework
- **PostgreSQL** – Relational database
- **Docker** – Containerization
- **Docker Compose** – Multi-container management
- **Nginx** – Reverse proxy

## Project Structure

```text
CSE-312-PROJ/
├── app/
│   ├── main.go
│   └── templates/
│       └── index.html
├── nginx/
│   └── default.conf
├── .dockerignore
├── .gitignore
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

Install the following before running the project:

- Docker Desktop
- Git

## Setup Instructions

### 1. Clone the repository

```bash
git clone YOUR_GITHUB_REPOSITORY_URL
cd CSE-312-PROJ
```

### 2. Configure environment variables

Create a `.env` file in the project root using `.env.example` as a template.

On Windows Command Prompt:

```cmd
copy .env.example .env
```

The `.env` file contains the PostgreSQL configuration. Do not commit this file to Git.

### 3. Build and start the application

Run:

```bash
docker compose up --build
```

Docker Compose starts three services:

- Go application
- PostgreSQL database
- Nginx reverse proxy

### 4. Open the application

Visit:

http://localhost:8080

### 5. Test the database connection

Visit:

http://localhost:8080/health

A successful response indicates that the application is running and connected to PostgreSQL.

Example response:

```json
{
  "status": "healthy",
  "database": "connected",
  "database_time": "2026-11-21T00:00:00Z"
}
```

The database time will vary.

## Architecture

The application uses three Docker services.

1. **Nginx:** Receives incoming HTTP requests on port 80 and forwards them to the Go application.
2. **Go application:** Handles application routes and communicates with PostgreSQL.
3. **PostgreSQL:** Stores application data and is accessible internally through the Docker network.

The application is exposed locally through Nginx on port 8080. PostgreSQL does not publish a host port.

## Stopping the Application

Press `Ctrl + C` in the terminal running Docker Compose, then run:

```bash
docker compose down
```

To remove the containers and database volume:

```bash
docker compose down -v
```

**Warning:** Removing the volume deletes the persisted PostgreSQL data.

## Deployment

Public deployment URL: **To be added after deployment.**