package server

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RequireNATS           bool
	NATSUrl               string
	FirebaseCredsFile     string
	MongoDBConnUrl        string
	FirebaseStorageBucket string
}

func ReadConfigFromEnv() Config {
	cfg := Config{RequireNATS: false}

	godotenv.Load()
	natsUrl, ok := os.LookupEnv("NATS_URL") // optional
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
	cfg.NATSUrl = natsUrl
	cfg.FirebaseCredsFile = credsFileName
	cfg.MongoDBConnUrl = mongodbConnectUrl
	cfg.FirebaseStorageBucket = storageBucket

	return cfg
}
