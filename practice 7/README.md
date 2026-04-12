# Practice 7

Решение собрано как небольшой JWT API на Go с `Gin`, `bcrypt` и in-memory хранилищем, чтобы проект можно было запустить без PostgreSQL и сразу показать все обязательные сценарии из задания.

## Что сделано

- `POST /users/` регистрирует пользователя и сохраняет пароль в хешированном виде.
- `POST /users/login` выдает JWT токен по логину и паролю.
- `GET /users/me` работает только через `JWTAuthMiddleware()` и определяет пользователя только по токену.
- `PATCH /users/promote/:id` защищен `RoleMiddleware("admin")`, поэтому доступен только администратору.
- Глобальный rate limiter использует:
  - `userID` из JWT для авторизованных запросов;
  - `ClientIP` для анонимных запросов.
- В rate limiter используется `sync.Mutex`, как требует задание.

## Как запустить

1. Установите Go.
2. В корне проекта выполните:

```powershell
go mod tidy
go run ./cmd/server
```

Если переменная `JWT_SECRET` не задана, сервер использует локальный dev-secret. При первом запуске автоматически создается bootstrap-админ:

- username: `admin`
- password: `admin123`
- email: `admin@example.com`

При необходимости можно переопределить через переменные окружения:

```powershell
$env:JWT_SECRET="super-secret"
$env:BOOTSTRAP_ADMIN_USERNAME="root"
$env:BOOTSTRAP_ADMIN_EMAIL="root@example.com"
$env:BOOTSTRAP_ADMIN_PASSWORD="root12345"
go run ./cmd/server
```

## Сценарий для демонстрации

### 1. Зарегистрировать двух обычных пользователей

```http
POST /users/
Content-Type: application/json

{
  "username": "user-a",
  "email": "a@example.com",
  "password": "secret123"
}
```

```http
POST /users/
Content-Type: application/json

{
  "username": "user-b",
  "email": "b@example.com",
  "password": "secret123"
}
```

### 2. Залогинить обоих пользователей

```http
POST /users/login
Content-Type: application/json

{
  "username": "user-a",
  "password": "secret123"
}
```

```http
POST /users/login
Content-Type: application/json

{
  "username": "user-b",
  "password": "secret123"
}
```

### 3. Показать, что `/users/me` зависит только от JWT

Передайте токен User A:

```http
GET /users/me
Authorization: Bearer <token-user-a>
```

В ответе должен быть email `a@example.com`.

Передайте токен User B:

```http
GET /users/me
Authorization: Bearer <token-user-b>
```

В ответе должен быть email `b@example.com`.

### 4. Показать role-based access

Сначала залогиньте bootstrap-админа:

```http
POST /users/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

Потом вызовите:

```http
PATCH /users/promote/<user-id>
Authorization: Bearer <admin-token>
```

Если тот же запрос сделать токеном обычного пользователя, сервер вернет `403 Forbidden`.

### 5. Показать rate limiter

Несколько раз подряд вызовите любой endpoint, например:

```http
GET /health
```

После превышения лимита сервер вернет:

```http
429 Too Many Requests
```

По умолчанию лимит: `5` запросов за `60` секунд. Его можно поменять через:

- `RATE_LIMIT_MAX_REQUESTS`
- `RATE_LIMIT_WINDOW_SECONDS`
