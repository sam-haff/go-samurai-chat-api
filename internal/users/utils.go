package users

import (
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetUserData(ctx *gin.Context, id string, data *accounts.UserData) bool {
	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)
	usersCollection := mongoInst.Collection(database.UsersCollection)

	filter := bson.D{{Key: "_id", Value: id}}

	res := usersCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		comm.AbortBadRequest(ctx, "Invalid user or user is not correctly registered", comm.CodeUserNotRegistered)
		return false
	}
	if data != nil {
		err := res.Decode(data)
		if err != nil {
			comm.AbortBadRequest(ctx, "Cant decode data from db", comm.CodeInvalidArgs)
			return false
		}
	}

	return true
}
