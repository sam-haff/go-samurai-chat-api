package main

import (
	"context"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/server"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	godotenv.Load()

	credsFileName, ok := os.LookupEnv("FIREBASE_CREDS_FILE")
	if !ok {
		log.Fatal("Service account is required to be set through env var file path to creds file")
	}

	mongodbConnectUrl, ok := os.LookupEnv("MONGODB_CONNECT_URL")
	if !ok {
		log.Fatal("Mongodb connection url with creds should be set through env file")
	}

	storageBucket, ok := os.LookupEnv("FIREBASE_STORAGE_BUCKET")
	if !ok {
		log.Fatal("Firebase storage bucket name should be set through env file")
	}

	opt := option.WithCredentialsFile(credsFileName)
	fbApp, err := firebase.NewApp(
		context.Background(),
		&firebase.Config{
			StorageBucket: storageBucket,
		},
		opt)
	if err != nil {
		log.Fatal("Failed to create Firebase app: " + err.Error())
	}

	mongoInst, err := database.NewMongoDBInstance(mongodbConnectUrl)
	if err != nil {
		log.Fatal("Failed to connect to mongo db: " + err.Error())
	}

	if err := server.Run(":8080", fbApp, mongoInst); err != nil {
		log.Fatal("Failure at running a server: " + err.Error())
	}
}
