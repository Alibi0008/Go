# Practice 9

Готовый проект по Go:

1. `Resilient HTTP Client`
2. `Loan Repayment / Idempotency`


## Что внутри

- `main.go` — выбор сценария запуска
- `retry.go` — retry, exponential backoff, full jitter, `httptest`
- `idempotency.go` — idempotency middleware, Redis storage, конкурентная симуляция
- `retry_test.go` — тесты retry-механизма
- `idempotency_test.go` — тесты middleware и повторной доставки ответа
- `docker-compose.yml` — локальный Redis для второй части


## Быстрый старт

### 1. Поднять Redis

```bash
docker compose up -d
```

### 2. Установить зависимости

```bash
go mod tidy
```

### 3. Запустить первую задачу

```bash
go run . -task=retry
```

### 4. Запустить вторую задачу

```bash
go run . -task=idempotency
```

Если Redis слушает другой адрес:

```bash
go run . -task=idempotency -redis-addr=localhost:6379 -redis-password= -redis-db=0
```

## Тесты

```bash
go test ./...
```