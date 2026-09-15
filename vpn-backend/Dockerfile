FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /app/server ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates wireguard-tools iptables

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]