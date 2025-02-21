# Builder stage
FROM golang:1.23
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --parseDependency --parseInternal --parseDepth 1

RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server

EXPOSE 8080

CMD ["/api-server"]
