package database

import "fmt"

func NewTestMongoDBInstance() (*MongoDBInstance, error) {
	//adminUsername := "root"
	//adminPassword := "root"
	connectURI := "mongodb://127.0.0.1:27017/db?replicaSet=rs0&retryWrites=true&w=majority"
	//connectURI := "mongodb://127.0.0.1:27017" ///?authSource=admin&replicaSet=rs0" //, adminUsername, adminPassword)
	//connectURI := "mongodb://127.0.0.1:27017/?replicaSet=rs0&directConnection=true" //, adminUsername, adminPassword)

	fmt.Printf("Trying to connect to MongoDB with connection uri %s \n", connectURI)

	return NewMongoDBInstance(connectURI)
}
