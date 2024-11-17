package accounts

import (
	"context"
	"fmt"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"log"
)

func SetupTestingAccounts(accs []auth.TestingAccount) {
	mongoInst, err := database.NewTestMongoDBInstance()

	if err != nil {
		log.Fatalf("Failed to connect to test database with %s", err.Error())
	}
	fmt.Print("Connected to mongodb!")

	ctx := context.Background()

	for _, acc := range accs {
		CreateDBUserRecordsInternal(ctx, mongoInst, acc.Uid, acc.Username, acc.Email)
	}
}
