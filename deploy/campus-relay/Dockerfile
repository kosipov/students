# Релей для кампуса как статический бинарник: образ получается ~10 МБ, без базовой системы.
FROM golang:1.23-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod ./
COPY deploy/campus-relay/main.go deploy/campus-relay/main.go
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /campus-relay ./deploy/campus-relay

FROM scratch
# HTTPS до кампуса требует корневые сертификаты; в scratch их нет.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /campus-relay /campus-relay
EXPOSE 8080
ENTRYPOINT ["/campus-relay"]
