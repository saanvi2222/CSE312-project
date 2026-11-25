FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Copy auth directory files into /app
COPY . .
RUN go build -o auth-server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auth-server .

EXPOSE 8081
CMD ["./auth-server"]