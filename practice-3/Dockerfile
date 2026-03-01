# ---------- Stage 1: Build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go


# ---------- Stage 2: Runtime ----------
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/database/migrations ./database/migrations

EXPOSE 8080

ENTRYPOINT ["./main"]