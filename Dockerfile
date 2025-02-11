# Stage 1: Build stage
FROM golang:1.23.2-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Stage 2: Final stage
FROM alpine:edge AS final

LABEL org.opencontainers.image.source=https://github.com/teamyapchat/yapchat-server
LABEL org.opencontainers.image.description="Backend for YapChat"
LABEL org.opencontainers.image.licenses=GPLv3

WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
ENTRYPOINT ["/app/server"]
