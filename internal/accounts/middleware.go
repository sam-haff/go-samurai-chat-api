package accounts

import (
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"

	"firebase.google.com/go/v4/storage"
	"github.com/gin-gonic/gin"
)

const (
	CtxVarUserData        = "user-userdata"
	CtxVarFirebaseStorage = "fb-storage"
)

// TODO: move to server/middleware?
func InjectStorage(fbStorage *storage.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarFirebaseStorage, fbStorage)
	}
}

func CompleteRegisteredMiddleware(ctx *gin.Context) {
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	userId := ctx.MustGet(auth.CtxVarUserId).(string)

	userData := UserData{}

	if !DBUserRegisterCompletedUtil(ctx, mongoInst, userId, &userData) {
		comm.AbortUnauthorized(ctx, "User register is incomplete, action is not authorized", comm.CodeUserNotRegistered)
		return
	}

	ctx.Set(CtxVarUserData, userData) // TODO: consider using pointer
}
