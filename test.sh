#!/bin/bash

HOST=localhost
PASS=0
FAIL=0
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "=================================="
echo "  Тестирование Subscription API"
echo "=================================="

check() {
    local desc="$1"
    local expected="$2"
    local method="$3"
    local url="$4"
    local data="$5"

    if [ -n "$data" ]; then
        status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" -d "$data")
    else
        status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url")
    fi

    if [ "$status" = "$expected" ]; then
        echo -e "${GREEN}[PASS]${NC} $desc"
        PASS=$((PASS+1))
    else
        echo -e "${RED}[FAIL]${NC} $desc (ожидался $expected, получен $status)"
        FAIL=$((FAIL+1))
    fi
}

echo ""
echo "--- Создание подписок ---"

check "Создать Yandex Plus" "201" "POST" "http://$HOST:8080/subscriptions" \
  '{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"04-2026"}'

check "Создать Netflix" "201" "POST" "http://$HOST:8080/subscriptions" \
  '{"service_name":"Netflix","price":600,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"01-2026","end_date":"12-2026"}'

echo ""
echo "--- Получение списка ---"

check "Список подписок" "200" "GET" "http://$HOST:8080/subscriptions?limit=10&offset=0" ""

echo ""
echo "--- Отчёты ---"

check "Отчёт: все подписки пользователя" "200" "GET" "http://$HOST:8080/subscriptions/report" \
  '{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"01-2026","end_date":"12-2026"}'

check "Отчёт: фильтр Yandex" "200" "GET" "http://$HOST:8080/subscriptions/report" \
  '{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","service_name":"Yandex","start_date":"01-2026","end_date":"12-2026"}'

check "Отчёт: фильтр Netflix" "200" "GET" "http://$HOST:8080/subscriptions/report" \
  '{"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","service_name":"Netflix","start_date":"01-2026","end_date":"12-2026"}'

check "Отчёт: без фильтров" "200" "GET" "http://$HOST:8080/subscriptions/report" \
  '{"start_date":"01-2026","end_date":"12-2026"}'

echo ""
echo "=================================="
echo -e "  Итого: ${GREEN}$PASS пройдено${NC}, ${RED}$FAIL провалено${NC}"
echo "=================================="