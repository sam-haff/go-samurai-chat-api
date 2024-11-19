package auth

import (
	"fmt"
	"go-chat-app-api/internal/comm"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxVarUserId       = "user-id"
	CtxVarAuthToken    = "auth-token"
	CtxVarFirebaseAuth = "fb-auth"
)

func InjectAuth(fbAuth Auth) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarFirebaseAuth, fbAuth)
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

	fbAuth := ctx.MustGet(CtxVarFirebaseAuth).(Auth)
	authToken, err := fbAuth.VerifyToken(ctx, authComps[1])

	if err != nil {
		fmt.Printf("Unauthorized with: %s \n", err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, comm.NewApiResponse("Invalid creds", comm.CodeNotAuthenticated))

		ctx.Set(CtxVarUserId, "")
		return
	}

	ctx.Set(CtxVarUserId, authToken.UID)
	ctx.Set(CtxVarAuthToken, authToken)
}
