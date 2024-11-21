package messages

import (
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setupPckgTestingAccounts()

	os.Exit(m.Run())
}

const pckgPrefix = "messages"

const (
	TestingAccountsInDBStartingIndex = 0
	TestingAccountsInDBCount         = 2
)

func getPckgTestingAuthRecords() []auth.TestingAuthRecord {
	return auth.GetTestingAuthRecords(pckgPrefix, TestingAccountsInDBStartingIndex, TestingAccountsInDBCount)
}
func getPckgTestingAccountsInfo() []accounts.TestingAccount {
	return accounts.GetTestingAccountsInfo(
		pckgPrefix,
		TestingAccountsInDBStartingIndex,
		TestingAccountsInDBCount,
	)
}
func setupPckgTestingAccounts() {
	accs := getPckgTestingAccountsInfo()

	accounts.SetupTestingAccounts(accs)
}
func setupPckgAuthMock(finalizeSetup bool) *auth.MockFbAuth {
	authMock := auth.SetupAuthMock(pckgPrefix, getPckgTestingAuthRecords(), finalizeSetup)

	return authMock
}
