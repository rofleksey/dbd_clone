FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /master-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /master-server /master-server
COPY migrations /migrations

EXPOSE 8080

CMD ["/master-server"]
