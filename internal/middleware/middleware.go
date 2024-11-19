package middleware

import (
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

func InjectFBApp(fbApp *firebase.App) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarFirebaseApp, fbApp)
	}
}
