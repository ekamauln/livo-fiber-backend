# Livotech Warehouse Management System (Go + Fiber + Gorm + PostgreSQL)

This is an Backend REST API service for Livotech Warehouse Management System built with Go, Fiber, Gorm, and PostgreSQL. This application provides a robust and efficient way to manage warehouse operations, including user management with role base action control, store management, channel management, product management, expedition management, package tracking, order processing, return processing, complain processing, and reporting. Below are the instructions to set up the development environment and run the application.

## Recommended IDE Setup

- [VS Code](https://code.visualstudio.com/) + [Go - Official](https://marketplace.visualstudio.com/items?itemName=golang.Go)

## Quick Start

```bash
# Clone the repo
git clone https://github.com/ekamauln/livo-fiber-backend.git

# Install dependencies
cd livo-fiber-backend
go mod tidy

# Env variables
Create a `.env` file in the root directory and add the following variables:

Database configuration:
  -DB_HOST=your_db_host
  -DB_PORT=your_db_port
  -DB_USER=your_db_user
  -DB_PASSWORD=your_db_password
  -DB_NAME=livo_fiber_db
  -DB_SSLMODE=disable
  -DB_TZ=your timezone

App configuration:
  -ENV=development/production
  -PORT=your_app_port
  -APP_URL=http://your_app_url
  -APP_NAME=Livotech Warehouse Management System API
  -LOG_LEVEL=your_log_level

Token configuration:
We are using PASETO for token management. Add the following variables.
  -PASETO_SYMMETRIC_KEY=your_32_byte_symmetric_key (can be generated using generate_key.go script at ./cmd)
  -ACCESS_TOKEN_TTL=your_access_token_ttl_in_minutes
  -REFRESH_TOKEN_TTL=your_refresh_token_ttl_in_days

CORS configuration:
  -CORS_ORIGINS=your_cors_origins (comma separated values)

DeepFace configuration:
We are using DeepFace for face recognition service integration. Add the following variable.
  -DEEPFACE_URL=http://your_deepface_service_url

# Run the app in development mode
go run main.go

# Build the app for production
go build -o livotech-app main.go

# Run the app in production mode
./livotech-app
```
