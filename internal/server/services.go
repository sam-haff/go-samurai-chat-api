package server

import (
	"context"
	"errors"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/storage"
	"github.com/nats-io/nats.go"
	"google.golang.org/api/option"
)

type Services struct {
	FirebaseApp     *firebase.App
	FirebaseStorage *storage.Client
	FirebaseAuth    auth.Auth
	MongoDB         *database.MongoDBInstance
	Nats            *nats.Conn
}

func (services *Services) Init(cfg Config) error {
	err := error(nil)
	if cfg.RequireNATS {
		log.Println("Nats required")
		if cfg.NATSUrl == "" {
			return errors.New("NATSUrl was required but none found")
		}
		services.Nats, err = nats.Connect(cfg.NATSUrl)
		log.Println(cfg.NATSUrl)
		if err != nil {
			log.Printf("%s\n", err.Error())
			return err
		}
	}
	services.FirebaseApp, err = InitFirebase(cfg)
	if err != nil {
		return err
	}
	services.MongoDB, err = database.NewMongoDBInstance(cfg.MongoDBConnUrl)
	if err != nil {
		return err
	}
	services.FirebaseAuth = auth.NewAuth(services.FirebaseApp)
	services.FirebaseStorage, err = services.FirebaseApp.Storage(context.TODO())

	return err
}

func InitFirebase(cfg Config) (*firebase.App, error) {
	opt := option.WithCredentialsFile(cfg.FirebaseCredsFile)
	fbApp, err := firebase.NewApp(
		context.Background(),
		&firebase.Config{
			StorageBucket: cfg.FirebaseStorageBucket,
		},
		opt)
	return fbApp, err
}
