package presence

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) {
	publicRoutes.GET("/online/:uid", handleIsOnline) // make path parametric(/uid)
}

func handleIsOnline(ctx *gin.Context) {
	// TODO: filter by ip or move to a private network
	pwd := os.Getenv("PRESENCE_PWD")
	log.Printf("PWD: %s\n", pwd)
	if ctx.GetHeader("Authorization") != pwd {
		log.Printf("Got PWD: %s\n", ctx.GetHeader("Authorization"))
		ctx.AbortWithStatus(404)
		return
	}

	presenceState := ctx.MustGet(CtxVarPresence).(*State)
	uid := ctx.Param("uid")

	if uid == "" {
		ctx.AbortWithStatus(400)
		return
	}

	isOnline := presenceState.IsOnline(uid)
	log.Printf("UID: %s; Online: %d\n", uid, b2i(isOnline))
	res := []byte{byte(b2i(isOnline))}

	ctx.Writer.Write(res)
	ctx.AbortWithStatus(200)
}
