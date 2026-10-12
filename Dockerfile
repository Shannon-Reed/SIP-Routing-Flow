
# Stage 1: Build the Go application
FROM golang:1.27.2-alpine AS builder

WORKDIR /app

# Download dependencies separately for build caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and compile the application
COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /app/server .

# Stage 2: Create a smaller runtime image
FROM alpine:3.22

# Install certificates for secure PostgreSQL connections
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only the compiled application
COPY --from=builder /app/server ./server

# Run the application as a non-root user
USER 10001:10001

EXPOSE 8080

CMD ["./server"]
