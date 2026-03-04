FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /game-server .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /game-server /game-server

CMD ["/game-server"]
