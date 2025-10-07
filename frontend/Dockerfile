FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod ./


RUN go mod tidy && \
    go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

COPY index.html .
COPY login.html .
COPY login.js .
COPY app.js .
COPY style.css .
COPY cart.js .
COPY cart.html .

EXPOSE 8089
CMD ["./main"]