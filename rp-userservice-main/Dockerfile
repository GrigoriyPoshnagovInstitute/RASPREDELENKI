FROM golang:1.25.3-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /userservice ./cmd/userservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /userservice /app/userservice

ENTRYPOINT ["/app/userservice"]