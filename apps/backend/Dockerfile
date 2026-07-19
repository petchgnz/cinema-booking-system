# Stage 1: Build
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/main.go
RUN go build -o seed ./cmd/seed/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/seed .

EXPOSE 8080

CMD ["./main"]