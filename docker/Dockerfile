FROM golang:alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o goe-report main.go

# Stage 2: Create the minimal Alpine image with cron
FROM alpine:latest

# Install tzdata for timezone support in cron
RUN apk --no-cache add tzdata ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/goe-report .

# Copy and set permission for the entrypoint script
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Define default environment variables. 
# You can override these via docker run -e or in a docker-compose file.
# Note: The application automatically pulls config from env variables prefixed with GOEREPORT_
# e.g. GOEREPORT_GOE_SERIAL=12345
ENV CRON_EXPRESSION="0 0 1 * *"
ENV CRON_COMMAND="/app/goe-report report --month \$(date -d \"@\$(( \$(date +%s) - 86400 ))\" +%m-%Y)"

# Execute the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
