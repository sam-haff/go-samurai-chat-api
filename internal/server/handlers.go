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
	publicRoutes.GET("/hi", func(ctx *gin.Context) {
		comm.GenericOK(ctx)
	})
}

func handleCheck(ctx *gin.Context) {
	fmt.Printf("Handle check... \n")

	authHeader := ctx.GetHeader("Authorization")
	authComps := strings.Split(authHeader, " ")
	if len(authComps) != 2 && authComps[0] != "Bearer" {
		fmt.Printf("Invalid header \n")
		ctx.String(400, "Invalid header")
		return
	}

	fbApp := ctx.MustGet(middleware.CtxVarFirebaseApp).(*firebase.App)
	fbAuth, _ := fbApp.Auth(ctx)
	_, err := fbAuth.VerifyIDToken(ctx, authComps[1])

	if err != nil {
		fmt.Printf("Unauthorized with %s \n", err.Error())
		ctx.String(401, "Unauthorized")
		return
	}

	ctx.String(200, "Authorized")
}
