package users

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"
)

func RegisterHandlers(authRouterGroup *gin.RouterGroup, publicRouterGroup *gin.RouterGroup) { //routers *gin.Engine) {
	authRouterGroup.GET("/users/id/:uid", handleGetUser)
	authRouterGroup.GET("/uid/:username", handleGetUid)
}

func handleGetUser(ctx *gin.Context) {
	targetUserId := ctx.Param("uid")

	userData := accounts.UserData{}
	if !GetUserData(ctx, targetUserId, &userData) {
		return
	}

	comm.GenericOKJSON(ctx, userData)
}
func handleGetUid(ctx *gin.Context) {
	targetUsername := ctx.Param("username")

	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)
	usernamesCollection := mongoInst.Collection(database.UsernamesCollection)

	usernameData := accounts.UsernameData{}

	filter := bson.D{{Key: "_id", Value: targetUsername}}
	res := usernamesCollection.FindOne(ctx, filter)

	if res.Err() != nil {
		comm.AbortBadRequest(ctx, "No such user", comm.CodeUserNotRegistered)
		return
	}

	err := res.Decode(&usernameData)
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to decode server response", comm.CodeUserNotRegistered)
		return
	}

	comm.GenericOKJSON(ctx, usernameData)
}
