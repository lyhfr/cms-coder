FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cmscoder-user-service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cmscoder-user-service .
COPY --from=builder /app/manifest/config /app/manifest/config
EXPOSE 39011
CMD ["./cmscoder-user-service"]
