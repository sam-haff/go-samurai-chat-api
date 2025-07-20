FROM golang:alpine

WORKDIR /app

COPY ./../../internal ./internal
COPY ./../../cmd ./cmd

COPY ./../../go.mod ./
COPY ./../../go.sum ./

RUN go build -o chat-api ./cmd/api_service

COPY ./../../firebase-adminsdk-config.json ./
COPY ./../../.env ./

EXPOSE 8080

CMD ["./chat-api"]