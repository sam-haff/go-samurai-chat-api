FROM golang:alpine

WORKDIR /app

COPY ../../internal ./internal
COPY ../../cmd ./cmd
COPY ../../firebase-adminsdk-config.json ./
COPY ../../.env ./
COPY ../../go.mod ./
COPY ../../go.sum ./

RUN go build -o chat-presence ./cmd/presence_service

EXPOSE 8080

CMD ["./chat-presence"]