FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

WORKDIR /app/cmd/payment

RUN CGO_ENABLED=0 GOOS=linux go build -o payment .

FROM scratch

COPY --from=builder /app/cmd/payment/payment /

CMD ["/payment"]