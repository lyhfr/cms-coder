FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o cmscoder-web-server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cmscoder-web-server .
COPY --from=builder /app/manifest/config /app/manifest/config
EXPOSE 39010
CMD ["./cmscoder-web-server"]
