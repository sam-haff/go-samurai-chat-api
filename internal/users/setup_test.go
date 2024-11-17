package users

import (
	"context"
	"errors"
	"fmt"
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"log"
	"os"
	"testing"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/mock"
)

// assumes running clean mongodb instance(e.g container)
func TestMain(m *testing.M) {
	SetupTestingAccounts()

	os.Exit(m.Run())
}

type TestingAccount struct {
	username string
	email    string
	token    string
	uid      string
}

func GetTestingAccountsInfo() []TestingAccount {
	return []TestingAccount{
		{"userspckg_testingacc1", "userpckg_testingacc1@t.com", "abcde", "userspckg-uid1"},
		{"userspckg_testingacc2", "userpckg_testingacc2@t.com", "ghjkl", "userspckg-uid2"},
	}
}

func SetupAuthMock() *auth.MockFbAuth {
	accs := GetTestingAccountsInfo()

	mockAuth := auth.MockFbAuth{}
	for _, acc := range accs {
		authToken := &fbauth.Token{}
		authToken.UID = acc.uid
		authToken.Firebase.SignInProvider = "email"
		mockAuth.On("VerifyToken", mock.Anything, acc.token).Return(authToken, nil)
	}

	mockAuth.On("VerifyToken", mock.Anything, mock.Anything).Return(nil, errors.New("Error placeholder for invalid creds"))

	return &mockAuth
}

func SetupTestingAccounts() {
	accs := GetTestingAccountsInfo()

	mongoInst, err := database.NewTestMongoDBInstance()

	if err != nil {
		log.Fatalf("Failed to connect to test database with %s", err.Error())
	}
	fmt.Print("Connected to mongodb!")

	ctx := context.Background()

	for _, acc := range accs {
		accounts.CreateDBUserRecordsInternal(ctx, mongoInst, acc.uid, acc.username, acc.email)
	}
}
