# Stage 1: Build stage
FROM --platform=$BUILDPLATFORM golang:1.23.2-alpine AS build
WORKDIR /app

# Use Docker's target architecture
ARG TARGETOS
ARG TARGETARCH

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for the target platform
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o server main.go

# Stage 2: Final stage
FROM --platform=$TARGETPLATFORM alpine:edge AS final
WORKDIR /app

LABEL org.opencontainers.image.source=https://github.com/teamyapchat/yapchat-server
LABEL org.opencontainers.image.description="Backend for YapChat"
LABEL org.opencontainers.image.licenses=GPLv3

# Copy binary from build stage
COPY --from=build /app/server .

# Set permissions (optional for execution issues)
RUN chmod +x /app/server

# Expose the application port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/server"]
