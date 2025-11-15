FROM golang:1.25.3-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /paymentservice ./cmd/paymentservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /paymentservice /app/paymentservice

ENTRYPOINT ["/app/paymentservice"]