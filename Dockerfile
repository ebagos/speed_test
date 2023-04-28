FROM golang:1.20.3-alpine3.17 AS builder

WORKDIR /app
COPY ./ ./
RUN go build -ldflags="-s -w" -o ./server

FROM alpine:3.17

WORKDIR /app
COPY --from=builder /app/server /app/server
ENTRYPOINT ["/app/server"]
EXPOSE 50051