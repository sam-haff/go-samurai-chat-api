package accounts

import (
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"log"

	"github.com/gin-gonic/gin"
)

const (
	CtxVarUserUsername = "user-username"
	CtxVarUserEmail    = "user-email"
)

func CompleteRegisteredMiddleware(ctx *gin.Context) {
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	//ctx.GetString()
	userId := ctx.MustGet(auth.CtxVarUserId).(string)

	userData := UserData{}
	usernameData := UsernameData{}

	log.Printf("mw registered %s \n", userId)

	if !DBUserRegisterCompletedUtil(ctx, mongoInst, userId, &userData, &usernameData) {
		comm.AbortUnauthorized(ctx, "User register is incomplete, action is not authorized", comm.CodeUserNotRegistered)
		return
	}

	ctx.Set(CtxVarUserUsername, userData.Username)
	ctx.Set(CtxVarUserEmail, userData.Email)
}
