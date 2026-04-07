# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o sshport .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/sshport .

# SSH host key is mounted at runtime — see docker-compose.yml
# This avoids baking private keys into the image

EXPOSE 22

ENTRYPOINT ["./sshport"]
