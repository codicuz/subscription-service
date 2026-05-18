#!/bin/bash

podman pod create \
  --name subscription-pod \
  --publish 8080:8080 \
  --publish 5432:5432

podman run -d \
  --pod subscription-pod \
  --name subscriptions_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=subscriptions \
  -v "$(pwd)/migrations/001_init.sql:/docker-entrypoint-initdb.d/001_init.sql:Z" \
  -v subscription-data:/var/lib/postgresql/data:Z \
  docker.io/library/postgres:16-alpine

echo "Ждём PostgreSQL..."
until podman exec subscriptions_db pg_isready -U postgres -d subscriptions 2>/dev/null; do
  sleep 2
done
echo "PostgreSQL готов"

podman build -f Dockerfile-podman -t subscription-app .

podman run -d \
  --pod subscription-pod \
  --name subscriptions_app \
  -e SERVER_PORT=8080 \
  -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=subscriptions \
  -e DB_SSLMODE=disable \
  -e LOG_LEVEL=info \
  subscription-app

echo "Готово! Сервис на http://localhost:8080"
