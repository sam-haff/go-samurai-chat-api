package accounts

import (
	"os"
	"testing"

	"go-chat-app-api/internal/auth"
)

func TestMain(m *testing.M) {
	setupPckgTestingAccounts()

	os.Exit(m.Run())
}

const pckgPrefix = "accounts"

func getPckgTestingAccountsInfo() []auth.TestingAccount {
	return auth.GetTestingAccountsInfo(pckgPrefix)
}
func setupPckgTestingAccounts() {
	accs := getPckgTestingAccountsInfo()

	SetupTestingAccounts(accs)
}
func setupPckgAuthMock() *auth.MockFbAuth {
	authMock := auth.SetupAuthMock(pckgPrefix)

	return authMock
}
