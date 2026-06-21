# Development stage — hot reload with Air
FROM golang:1.25-alpine AS dev
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

# Production stage — minimal image
FROM alpine:3.19 AS production
WORKDIR /app
COPY --from=builder /server ./server
COPY migrations/ ./migrations/
EXPOSE 8080
CMD ["./server"]
