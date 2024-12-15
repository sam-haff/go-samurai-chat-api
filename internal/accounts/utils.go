package accounts

import (
	"context"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UtilStatus int

const (
	UtilErrorOk          = UtilStatus(0)
	UtilErrorDoesntExist = UtilStatus(1)
	UtilErrorDecode      = UtilStatus(2)
)

func dbCreateUserRecordsInternal(ctx context.Context, mongoInst *database.MongoDBInstance, userData UserData) error {
	usersCollection := mongoInst.Collection(database.UsersCollection)

	_, err := usersCollection.InsertOne(ctx, userData)
	if err != nil {
		return err
	}

	return nil
}

func DBFindOneQueryUtil(ctx context.Context, mongoInst *database.MongoDBInstance, col *mongo.Collection, filter bson.D, data any) UtilStatus {
	res := col.FindOne(ctx, filter)

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
func DBGetUserDataByUsernameUtil(ctx context.Context, mongoInst *database.MongoDBInstance, username string, data *UserData) UtilStatus {
	usersCollection := mongoInst.Collection(database.UsersCollection)

	filter := bson.D{{Key: "username", Value: username}}

	return DBFindOneQueryUtil(ctx, mongoInst, usersCollection, filter, data)
}
func DBGetUserDataUtil(ctx context.Context, mongoInst *database.MongoDBInstance, id string, data *UserData) UtilStatus {
	usersCollection := mongoInst.Collection(database.UsersCollection)

	filter := bson.D{{Key: "_id", Value: id}}

	return DBFindOneQueryUtil(ctx, mongoInst, usersCollection, filter, data)
}

func DBGetUserDataByUsername(ctx *gin.Context, username string, data *UserData) bool {

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	utilErr := DBGetUserDataByUsernameUtil(ctx, mongoInst, username, data)

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
func DBUserRegisterCompletedUtil(ctx context.Context, mongoInst *database.MongoDBInstance, uid string, userData *UserData) bool {
	if userData == nil {
		panic("arguments should not be null")
		return false
	}
	status := DBGetUserDataUtil(ctx, mongoInst, uid, userData)
	if userData.Username == "" {
		return false
	}
	return status == UtilErrorOk
}
