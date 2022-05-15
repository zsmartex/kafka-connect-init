FROM golang:1.18.2-alpine3.14 AS builder

WORKDIR /build
ENV CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o kafka-connect-init main.go


FROM alpine:3

WORKDIR /app

COPY --from=builder /build/kafka-connect-init ./
