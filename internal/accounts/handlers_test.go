package accounts

import (
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func GetRoutes() *gin.Engine {
	authMock := setupPckgAuthMock()
	testMongoInst, _ := database.NewTestMongoDBInstance()

	routes := gin.Default()
	routes.Use(middleware.InjectParams(nil, authMock, testMongoInst))
	authRoutes := routes.Group("/", middleware.AuthMiddleware)
	publicRoutes := routes.Group("/")
	RegisterHandlers(authRoutes, publicRoutes)

	return routes
}
