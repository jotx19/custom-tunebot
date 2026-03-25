FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ffmpeg gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/bot

FROM alpine:latest
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app
COPY --from=builder /app/app .

CMD ["./app"]