FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build ./cmd/vaultix

FROM alpine:latest AS app
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/vaultix .
CMD ["./vaultix"]

FROM postgres:latest AS db
COPY ./migrations /docker-entrypoint-initdb.d/
