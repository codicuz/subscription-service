#!/bin/bash

HOST=localhost

curl -X POST http://$HOST:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "04-2026"
  }'

curl -X POST http://$HOST:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Netflix",
    "price": 600,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "01-2026",
    "end_date": "12-2026"
  }'

curl -X GET http://$HOST:8080/subscriptions/report \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "01-2026",
    "end_date": "12-2026"
  }'
