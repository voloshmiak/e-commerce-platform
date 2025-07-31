FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy && \
    go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/api

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8084
CMD ["./main"]