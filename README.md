# Subscription Service

## Запуск

### Docker Compose

```bash
docker compose up --build
```

### Podman

```bash
chmod +x podman-start.sh
./podman-start.sh
```

Сервис будет на `http://localhost:8080`.

## Тестирование

```bash
chmod +x test.sh
./test.sh
```

## Остановка

### Docker Compose

```bash
docker compose down
```

### Podman

```bash
podman pod stop subscription-pod
podman pod rm subscription-pod
podman volume rm subscription-data
```

## API

### Создать подписку
```
POST /subscriptions
{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "04-2026"
}
```

### Получить список
```
GET /subscriptions?limit=10&offset=0
```

### Получить по ID
```
GET /subscriptions/{id}
```

### Обновить
```
PUT /subscriptions/{id}
{
    "price": 500
}
```

### Удалить
```
DELETE /subscriptions/{id}
```

### Отчёт по стоимости
```
GET /subscriptions/report
{
    "user_id": "...",
    "service_name": "...",
    "start_date": "01-2026",
    "end_date": "12-2026"
}
```

## Swagger

http://localhost:8080/swagger/index.html

## Конфигурация (.env)

```
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=subscriptions
DB_SSLMODE=disable
LOG_LEVEL=info
```