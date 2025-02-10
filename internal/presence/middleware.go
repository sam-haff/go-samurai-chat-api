package presence

import "github.com/gin-gonic/gin"

const CtxVarPresence = "presence"

func InjectPresenceState(state *State) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarPresence, state)
	}
}
