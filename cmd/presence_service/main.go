package main

import (
	"go-chat-app-api/internal/presence"
	"go-chat-app-api/internal/server"
	"go-chat-app-api/internal/upload"
	"log"
)

func RegisterHandlers(server *server.Server) {
	upload.RegisterHandlers(server.AuthRoutes, server.PublicRoutes)
}

func main() {
	cfg := server.ReadConfigFromEnv()
	cfg.RequireNATS = true

	server, err := server.Setup(cfg)
	if err != nil {
		log.Fatalf("Failed to init server: %s", err.Error())
	}

	state := presence.NewState()
	server.PublicRoutes.Use(presence.InjectPresenceState(state))

	presence.RegisterHandlers(server.AuthRoutes, server.PublicRoutes)
	closeState := state.RegisterNatsListeners(server.Services.Nats)

	if err = server.Run(":8080"); err != nil {
		closeState()
		log.Fatalf("Failed to run server: %s", err.Error())
	}
	closeState()
}
