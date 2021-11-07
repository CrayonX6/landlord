FROM golang:1.14.7-alpine3.12 AS builder

WORKDIR /app

ENV GOPROXY https://goproxy.cn,direct
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o api cmd/api/main.go
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o queue cmd/queue/main.go

FROM alpine:3.12

WORKDIR /app
COPY --from=builder /app/api /app/
COPY --from=builder /app/queue /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY config.toml /app/

#CMD ["/app/api"]
#CMD ["/app/queue"]
COPY build/supervisor /etc/supervisor.d/