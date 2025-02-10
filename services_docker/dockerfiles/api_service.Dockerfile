FROM golang:alpine

WORKDIR /app

COPY ./../../internal ./internal
COPY ./../../cmd ./cmd
COPY ./../../firebase-adminsdk-config.json ./
COPY ./../../.env ./
COPY ./../../go.mod ./
COPY ./../../go.sum ./

RUN go build -o chat-api ./cmd/api_service

EXPOSE 8080

CMD ["./chat-api"]