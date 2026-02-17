#!/bin/sh
set -e
export DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

echo "Running database migrations..."
goose -dir ./migrations postgres "$DATABASE_URL" up

echo "Starting application..."
exec ./application