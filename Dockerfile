FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o go-chat-app-api ./cmd

EXPOSE 8080

CMD ["./go-chat-app-api"]