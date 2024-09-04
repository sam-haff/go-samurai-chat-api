FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build .

EXPOSE 8080

CMD ["./go-chat-app-api"]