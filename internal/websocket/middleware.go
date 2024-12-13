package websocket

import "github.com/gin-gonic/gin"

func InjectWsHub(hub *WsHub) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarHub, hub)
	}
}
