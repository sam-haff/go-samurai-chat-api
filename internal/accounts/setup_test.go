package accounts

import (
	"go-chat-app-api/internal/auth"
	"os"
	"testing"
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
	return auth.SetupAuthMock(pckgPrefix)
}
