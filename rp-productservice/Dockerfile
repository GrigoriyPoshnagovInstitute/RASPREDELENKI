FROM golang:1.25.3-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o /productservice ./cmd/productservice

FROM alpine:latest

WORKDIR /app

COPY --from=builder /productservice /app/productservice

ENTRYPOINT ["/app/productservice"]