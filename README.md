# Todo App

A modern, containerized Todo application built with Go, featuring a RESTful API backend and MySQL database. The application uses Echo framework for the HTTP server, GORM for database operations, and Zap for logging.

## Features & Tech Stack

- RESTful API backend built with Go and Echo framework
- MySQL database with GORM ORM
- Structured logging with Zap
- Hot-reload development environment with Air
- Docker containerization
- Makefile for easy development workflow

## Prerequisites

- Docker and Docker Compose
- Go 1.24.3 or later
- Make (optional, for using Makefile commands)

## Getting Started

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd todo-app
   ```

2. Start the application using Docker Compose:
   ```bash
   make start
   ```
   This will start both the application and MySQL database containers.

3. The application will be available at:
   - API Server: http://localhost:8765
   - MySQL Database: localhost:3306

## API Documentation

Detailed API documentation is available in each internal domain package.
It includes available endpoints, request / response formats, data models, error handling.

ToDo Items - [read here](internal/todos/API.md)

## Development

The project includes several Makefile commands to help with development:

- `make start` - Start the containers
- `make stop` - Stop the containers
- `make restart` - Restart the containers
- `make logs` - View container logs

To follow logs for a specific service:
```bash
make logs srv=todo-app
```