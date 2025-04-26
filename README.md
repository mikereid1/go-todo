# 📝 TODO App

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Everyone has to build a todo app at some point.
I never made one in Go, so now it's time to *go* and do that.

## 📋 Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Running Tests](#running-tests)

## 🎯 Overview

A RESTful Todo application that demonstrates the use of modern Go practices, event-driven architecture, and NoSQL
database integration.

## 🛠 Tech Stack

- Go 1.24
- Chi Router for HTTP routing
- MongoDB for data persistence
- RabbitMQ for event messaging
- Docker for containerization

## 🚀 Getting Started

1. Clone the repository
2. Install dependencies
3. Set up MongoDB and RabbitMQ
4. Configure environment variables
5. Run the application

### 🐳 Local Development with Docker/Podman

For local development, you can use either Docker Compose or Podman Compose to start all dependencies:

1. Make sure you have installed either:
    - Docker and Docker Compose, or
    - Podman and Podman Compose

2. Start the services:
   ```bash
   # Using Docker
   docker compose up -d

   # Using Podman
   podman-compose up -d
   ```

3. The following services will be available:
    - MongoDB:
        - Host: localhost
        - Port: 27017
        - Connection string: mongodb://localhost:27017
    - RabbitMQ:
        - AMQP Port: 5672
        - Management UI: http://localhost:15672
        - Default credentials: guest/guest

## 📡 API Endpoints

| Method | Endpoint           | Description         |
|--------|--------------------|---------------------|
| GET    | /api/v1/todos      | List all todos      |
| POST   | /api/v1/todos      | Create a todo       |
| GET    | /api/v1/todos/{id} | Get a specific todo |
| PUT    | /api/v1/todos/{id} | Update a todo       |
| DELETE | /api/v1/todos/{id} | Delete a todo       |

## 🧪 Running Tests

To run the tests, use the following commands:

### Unit Tests
```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests and generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```



