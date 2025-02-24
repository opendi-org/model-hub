# Builder stage
FROM golang:1.23
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --parseDependency --parseInternal --parseDepth 1

#hot reloading - DOES NOT WORK
#RUN go install github.com/air-verse/air@latest 
#RUN air init

#For actually building the API
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-server

EXPOSE 8080

#If building the API
CMD ["/api-server"]

#DOES NOT WORK
#for hot reloading
#CMD ["air", "-c", ".air.toml"]
