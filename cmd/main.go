package main

import (
	"context"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/server"
	"log"
	"net/mail"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func emailIsvalid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func main() {
	godotenv.Load()

	credsFileName, ok := os.LookupEnv("FIREBASE_CREDS_FILE")
	if !ok {
		log.Fatal("Service account is required to be set through env var file path to creds file")
	}
	mongodbConnectUrl, ok := os.LookupEnv("MONGODB_CONNECT_URL")
	if !ok {
		log.Fatal("Mongodb connection url with creds should be set thorugh env file")
	}

	opt := option.WithCredentialsFile(credsFileName)
	fbApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to create Firebase app: %s", err.Error())
	}

	mongoInst, err := database.NewMongoDBInstance(mongodbConnectUrl)
	if err != nil {
		log.Fatalf("Failed to connect to mongo db: %s", err.Error())
	}

	if err := server.Run(":8080", fbApp, mongoInst); err != nil {
		log.Fatalf("Failure at running a server: %s", err.Error())
	}
}
