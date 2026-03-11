# Stage 1: Build from source (for local builds)
FROM golang:alpine AS local-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o goe-report main.go

# Stage 2: Base production environment
FROM alpine:latest AS release-base

# Install tzdata for timezone support in cron
RUN apk --no-cache add tzdata ca-certificates

WORKDIR /app

# Copy and set permission for the entrypoint script
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Define default environment variables. 
# You can override these via docker run -e or in a docker-compose file.
# Note: The application automatically pulls config from env variables prefixed with GOEREPORT_
# e.g. GOEREPORT_GOE_SERIAL=12345
ENV CRON_EXPRESSION="0 0 1 * *"
ENV CRON_COMMAND="/app/goe-report report --pdf --attach-pdfs --send-mail"

# Execute the entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]

# Stage 3: For GoReleaser (binary provided in context by GoReleaser)
FROM release-base AS goreleaser
# GoReleaser will copy the built binary into the docker build context natively
COPY goe-report /app/goe-report

# Stage 4: For local Docker build (binary copied from local-builder)
# This is the default stage when running `docker build` without a target
FROM release-base AS local
COPY --from=local-builder /app/goe-report /app/goe-report
