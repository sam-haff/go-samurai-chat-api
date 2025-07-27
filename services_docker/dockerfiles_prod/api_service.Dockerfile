FROM golang:alpine

WORKDIR /app

COPY ./../../go.mod ./
COPY ./../../go.sum ./
RUN go mod download

COPY ./../../internal ./internal
COPY ./../../cmd ./cmd

RUN go build -o chat-api ./cmd/api_service

EXPOSE 8080

CMD ["./chat-api"]