package database

import "fmt"

func NewTestMongoDBInstance() (*MongoDBInstance, error) {
	connectURI := "mongodb://127.0.0.1:27017/db?replicaSet=rs0&retryWrites=true&w=majority"

	fmt.Printf("Trying to connect to MongoDB with connection uri %s \n", connectURI)

	return NewMongoDBInstance(connectURI)
}
