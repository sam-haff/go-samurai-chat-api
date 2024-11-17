package server

import (
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/messages"
	"go-chat-app-api/internal/middleware"
	"go-chat-app-api/internal/users"
)

func Run(addr string, fbApp *firebase.App, mongoInst *database.MongoDBInstance) error {
	fbAuth := auth.NewAuth(fbApp)

	routers := gin.Default()

	routers.Use(middleware.InjectParams(fbApp, fbAuth, mongoInst))
	authRoutes := routers.Group("/", middleware.AuthMiddleware)
	publicRoutes := routers.Group("/")

	RegisterHandlers(authRoutes, publicRoutes)

	accounts.RegisterHandlers(authRoutes, publicRoutes)
	users.RegisterHandlers(authRoutes, publicRoutes)
	messages.RegisterHandlers(authRoutes, publicRoutes)

	routers.Run(addr)

	return nil
}
