// handling user files uploads
package main

import (
	"go-chat-app-api/internal/server"
	"go-chat-app-api/internal/upload"
	"log"
)

func RegisterHandlers(server *server.Server) {
	upload.RegisterHandlers(server.AuthRoutes, server.PublicRoutes)
}

func main() {
	cfg := server.ReadConfigFromEnv()
	cfg.RequireNATS = false

	server, err := server.Setup(cfg)
	if err != nil {
		log.Fatalf("Failed to init server: %s", err.Error())
	}

	RegisterHandlers(server)

	if err = server.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %s", err.Error())
	}
}
