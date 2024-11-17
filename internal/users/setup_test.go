package users

import (
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"os"
	"testing"
)

// assumes running clean mongodb instance(e.g container)
func TestMain(m *testing.M) {
	setupPckgTestingAccounts()

	os.Exit(m.Run())
}

const usersPckgPrefix = "userspckg"

func getPckgTestingAccountsInfo() []auth.TestingAccount {
	return auth.GetTestingAccountsInfo(usersPckgPrefix)
}

func setupPckgAuthMock() *auth.MockFbAuth {
	return auth.SetupAuthMock(usersPckgPrefix)
}

func setupPckgTestingAccounts() {
	accs := getPckgTestingAccountsInfo()

	accounts.SetupTestingAccounts(accs)
}
