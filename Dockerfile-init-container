# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy init-container source
COPY init-container/ ./

# Download dependencies (if any)
RUN go mod tidy

# Build the init-container binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o init-container .

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for health checks
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/init-container .

# Run the initialization
CMD ["./init-container"]