package server

import (
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/messages"
	"go-chat-app-api/internal/middleware"
)

func Run(addr string, fbApp *firebase.App, mongoInst *database.MongoDBInstance) error {
	fbAuth := auth.NewAuth(fbApp)

	routers := gin.Default()

	routers.Use(middleware.InjectFBApp(fbApp), auth.InjectAuth(fbAuth), database.InjectDB(mongoInst))
	authRoutes := routers.Group("/", auth.AuthMiddleware)
	publicRoutes := routers.Group("/")

	RegisterHandlers(authRoutes, publicRoutes)

	accounts.RegisterHandlers(authRoutes, publicRoutes)
	messages.RegisterHandlers(authRoutes, publicRoutes)

	return routers.Run(addr)
}
