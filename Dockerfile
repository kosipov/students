# syntax=docker/dockerfile:1
FROM node:12.18.1
WORKDIR /app
COPY . ./
RUN npm install --production
RUN npm run build

FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
COPY --from=0 /app/templates/dist ./templates/dist

RUN go build -o ./student-app ./cmd/api/main.go
EXPOSE 8080

CMD [ "./student-app" ]


