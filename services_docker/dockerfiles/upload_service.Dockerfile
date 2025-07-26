FROM golang:alpine

WORKDIR /app

COPY ./../../go.mod ./
COPY ./../../go.sum ./
RUN go mod download

COPY ../../internal ./internal
COPY ../../cmd ./cmd

RUN go build -o chat-upload ./cmd/upload_service

COPY ../../firebase-adminsdk-config.json ./
COPY ../../.env ./

EXPOSE 8080

CMD ["./chat-upload"]