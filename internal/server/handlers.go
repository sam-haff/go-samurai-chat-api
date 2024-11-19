package server

import (
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/middleware"
)

func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) { //routers *gin.Engine) {
}

// only for testing purposes
func handleCheck(ctx *gin.Context) {
	fmt.Printf("Handle check... \n")

	authHeader := ctx.GetHeader("Authorization")
	authComps := strings.Split(authHeader, " ")
	if len(authComps) != 2 && authComps[0] != "Bearer" {
		fmt.Printf("Invalid header \n")
		comm.AbortBadRequest(ctx, "Invalid header", comm.CodeInvalidArgs)
		return
	}

	fbApp := ctx.MustGet(middleware.CtxVarFirebaseApp).(*firebase.App)
	fbAuth, _ := fbApp.Auth(ctx)
	_, err := fbAuth.VerifyIDToken(ctx, authComps[1])

	if err != nil {
		fmt.Printf("Unauthorized with %s \n", err.Error())
		comm.AbortUnauthorized(ctx, "Unauthorized", comm.CodeNotAuthenticated)
		return
	}

	comm.GenericOK(ctx)
}
