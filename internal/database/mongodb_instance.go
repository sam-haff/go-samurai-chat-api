package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBInstance struct {
	Client *mongo.Client
}

func (db *MongoDBInstance) Database() *mongo.Database {
	return db.Client.Database(DatabaseName)
}
func (db *MongoDBInstance) Collection(name string) *mongo.Collection {
	return db.Database().Collection(name)
}

func NewMongoDBInstance(connectURI string) (*MongoDBInstance, error) {
	mongoServerAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectURI).SetServerAPIOptions(mongoServerAPI).SetConnectTimeout(3 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, opts)
	inst := MongoDBInstance{Client: mongoClient}

	return &inst, err
}
