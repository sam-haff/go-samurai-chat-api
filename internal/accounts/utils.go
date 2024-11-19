package accounts

import (
	"context"
	"fmt"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const (
	UtilErrorOk          = 0
	UtilErrorDoesntExist = 1
	UtilErrorDecode      = 2
)

type UtilStatus int

func dbCreateUserRecordsInternal(ctx context.Context, mongoInst *database.MongoDBInstance, uid string, username string, email string) error {
	usersCollection := mongoInst.Collection(database.UsersCollection)
	usernamesCollection := mongoInst.Collection(database.UsernamesCollection)

	wc := writeconcern.Majority()
	txOptions := options.Transaction().SetWriteConcern(wc)
	session, err := mongoInst.Client.StartSession()
	if err != nil {
		fmt.Printf("Failed to start session \n")
		return fmt.Errorf("Failed to start tx with %v", err)
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(
		ctx,
		func(ctx mongo.SessionContext) (interface{}, error) {
			_, err := usersCollection.InsertOne(ctx, UserData{
				Id:       uid,
				Username: username,
				Email:    email,
			})
			if err != nil {
				return nil, err
			}
			_, err = usernamesCollection.InsertOne(ctx, UsernameData{
				Id:     username,
				UserID: uid,
			})

			return nil, err
		},
		txOptions,
	)
	if err != nil {

		fmt.Printf("Tx failed with error: %s\n", err.Error())
		return fmt.Errorf("Failed to create db records with: %v", err)
	}

	return nil
}

func DBGetUsernameDataUtil(ctx context.Context, mongoInst *database.MongoDBInstance, username string, data *UsernameData) UtilStatus {
	usersCollection := mongoInst.Collection(database.UsernamesCollection)

	filter := bson.D{{Key: "_id", Value: username}}

	res := usersCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		return UtilErrorDoesntExist
	}
	if data != nil {
		err := res.Decode(data)
		if err != nil {
			return UtilErrorDecode
		}
	}

	return UtilErrorOk
}

func DBGetUserDataUtil(ctx context.Context, mongoInst *database.MongoDBInstance, id string, data *UserData) UtilStatus {
	usersCollection := mongoInst.Collection(database.UsersCollection)

	filter := bson.D{{Key: "_id", Value: id}}

	res := usersCollection.FindOne(ctx, filter)

	if res.Err() != nil {
		return UtilErrorDoesntExist
	}

	if data != nil {
		err := res.Decode(data)
		if err != nil {
			return UtilErrorDecode
		}

		return UtilErrorOk
	}

	return UtilErrorOk
}

func DBGetUsernameData(ctx *gin.Context, username string, data *UsernameData) bool {
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	status := DBGetUsernameDataUtil(ctx, mongoInst, username, data)

	if status == UtilErrorDoesntExist {
		comm.AbortBadRequest(ctx, "Invalid user or user is not correctly registered", comm.CodeUserNotRegistered)
		return false
	}
	if status == UtilErrorDecode {
		comm.AbortBadRequest(ctx, "Cant decode data from db", comm.CodeInvalidArgs)
		return false
	}

	return true
}

func DBGetUserData(ctx *gin.Context, id string, data *UserData) bool {

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	utilErr := DBGetUserDataUtil(ctx, mongoInst, id, data)

	if utilErr == UtilErrorDoesntExist {
		comm.AbortBadRequest(ctx, "Invalid user or user is not correctly registered", comm.CodeUserNotRegistered)
		return false
	}
	if utilErr == UtilErrorDecode {
		comm.AbortBadRequest(ctx, "Cant decode data from db", comm.CodeInvalidArgs)
		return false
	}

	return true
}
func DBUserRegisterCompletedUtil(ctx context.Context, mongoInst *database.MongoDBInstance, uid string, userData *UserData, usernameData *UsernameData) bool {
	if userData == nil || usernameData == nil {
		panic("arguments should not be null")
		return false
	}
	status := DBGetUserDataUtil(ctx, mongoInst, uid, userData)
	if status != UtilErrorOk {
		return false
	}
	status = DBGetUsernameDataUtil(ctx, mongoInst, userData.Username, usernameData)
	return status == UtilErrorOk
}
