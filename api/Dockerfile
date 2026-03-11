FROM golang:1.26-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /opendi-api ./cmd/api

EXPOSE 8080
CMD ["/opendi-api"]
