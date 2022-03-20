FROM golang:1.17.3-alpine3.14 AS builder

WORKDIR /build
ENV CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o kafka-connect-init main.go


FROM alpine:20220316

WORKDIR /app

COPY --from=builder /build/kafka-connect-init ./
