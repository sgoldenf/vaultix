FROM golang:latest AS app
WORKDIR /app
COPY . .        
RUN go mod download
RUN go build -o vaultix ./cmd/vaultix
COPY ./.env .
CMD ["./vaultix"]
