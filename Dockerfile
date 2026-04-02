FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o metricsink ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/metricsink .
EXPOSE 8080
CMD ["./metricsink"]
