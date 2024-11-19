package database

import "github.com/gin-gonic/gin"

const CtxVarMongoDBInst = "mongo-db"

func InjectDB(mongoInst *MongoDBInstance) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarMongoDBInst, mongoInst)
	}
}
