# Builder stage
FROM golang:1.23

ENV OPEN_DI_DB_NAME open_di_model_hub
ENV OPEN_DI_DB_PORT 3306
ENV OPEN_DI_DB_HOSTNAME db
ENV OPEN_DI_DB_PASSWORD temp_pass
ENV OPEN_DI_DB_USERNAME root
ENV OPENDI_MODEL_HUB_ADDRESS localhost
ENV OPENDI_MODEL_HUB_PORT 8080

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
