package accounts

import (
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"

	"github.com/gin-gonic/gin"
)

const (
	CtxVarUserUsername = "user-username"
	CtxVarUserEmail    = "user-email"
)

func CompleteRegisteredMiddleware(ctx *gin.Context) {
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	userId := ctx.MustGet(auth.CtxVarUserId).(string)
	if len(userId) == 0 {
		// shouldnt ever reach there
		comm.AbortUnauthorized(ctx, "Not authorized", comm.CodeNotAuthenticated)
		return
	}

	userData := UserData{}
	usernameData := UsernameData{}

	if !DBUserRegisterCompletedUtil(ctx, mongoInst, userId, &userData, &usernameData) {
		comm.AbortUnauthorized(ctx, "User register is incomplete, action is not authorized", comm.CodeUserNotRegistered)
	}

	ctx.Set(CtxVarUserUsername, userData.Username)
	ctx.Set(CtxVarUserEmail, userData.Email)
}
