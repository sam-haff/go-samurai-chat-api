package messages

import "github.com/gin-gonic/gin"

const CtxVarFcm = "fcm"

func InjectFcm(fcm FcmClient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarFcm, fcm)
	}
}
