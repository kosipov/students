# syntax=docker/dockerfile:1

# Frontend: the SPA from web/ built into web/dist
FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web ./
RUN npm run build

# Go binary: both DB drivers are pure Go, so the binary is fully static
FROM golang:1.19-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /student-app ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /create-admin ./cmd/createadmin

# Runtime: config and the frontend are resolved relative to the working directory
FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /student-app ./student-app
COPY --from=build /create-admin ./create-admin
COPY config ./config
COPY --from=frontend /web/dist ./web/dist

USER app
ENV GIN_MODE=release
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${PORT:-8080}/healthz" || exit 1

CMD ["./student-app"]
