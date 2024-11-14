package middleware

import (
	"fmt"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
)

func InjectParams(app *firebase.App, mongoInst *database.MongoDBInstance) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarFirebaseApp, app)
		ctx.Set(CtxVarMongoDBInst, mongoInst)
	}
}
func AuthMiddleware(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	authComps := strings.Split(authHeader, " ")
	if len(authComps) != 2 && authComps[0] != "Bearer" {
		fmt.Printf("Invalid header \n")

		ctx.AbortWithStatusJSON(http.StatusBadRequest, comm.NewApiResponse("Invalid header", comm.CodeNotAuthenticated))
		return
	}

	fbApp := ctx.MustGet(CtxVarFirebaseApp).(*firebase.App)
	fbAuth, _ := fbApp.Auth(ctx)
	authToken, err := fbAuth.VerifyIDToken(ctx, authComps[1])

	if err != nil {
		fmt.Printf("Unauthorized with: %s \n", err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, comm.NewApiResponse("Invalid creads", comm.CodeNotAuthenticated))

		ctx.Set(CtxVarUserId, "")
		return
	}

	ctx.Set(CtxVarUserId, authToken.UID) //TODO: remove
	ctx.Set(CtxVarAuthToken, authToken)
}
