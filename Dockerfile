FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add --no-cache ca-certificates
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Use scratch for a small runtime image — only needs the builder base (already cached).
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /app/server
COPY web /app/web

WORKDIR /app

EXPOSE 8080

CMD ["/app/server"]
