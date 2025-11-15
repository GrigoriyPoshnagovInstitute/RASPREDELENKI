FROM golang:1.25.3-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /notificationservice ./cmd/notificationservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /notificationservice /app/notificationservice

ENTRYPOINT ["/app/notificationservice"]