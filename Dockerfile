FROM golang:alpine

WORKDIR /app

COPY internal ./
COPY cmd ./
COPY firebase-adminsdk-config.json ./
COPY .env ./

RUN go build -o go-chat-app-api ./cmd

EXPOSE 8080

CMD ["./go-chat-app-api"]