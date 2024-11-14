package database

import (
	"context"

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
	opts := options.Client().ApplyURI(connectURI).SetServerAPIOptions(mongoServerAPI)
	mongoClient, err := mongo.Connect(context.TODO(), opts)
	inst := MongoDBInstance{Client: mongoClient}

	return &inst, err
}
