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

const (
	TestingAccountsInDBStartingIndex = 0
	TestingAccountsInDBCount         = 2
)

func getPckgTestingAuthRecords() []auth.TestingAuthRecord {
	return auth.GetTestingAuthRecords(pckgPrefix, TestingAccountsInDBStartingIndex, TestingAccountsInDBCount)
}
func getPckgTestingAccountsInfo() []TestingAccount {
	return GetTestingAccountsInfo(
		pckgPrefix,
		TestingAccountsInDBStartingIndex,
		TestingAccountsInDBCount,
	)
}
func setupPckgTestingAccounts() {
	accs := getPckgTestingAccountsInfo()

	SetupTestingAccounts(accs)
}
func setupPckgAuthMock(finalizeSetup bool) *auth.MockFbAuth {
	authMock := auth.SetupAuthMock(pckgPrefix, getPckgTestingAuthRecords(), finalizeSetup)

	return authMock
}
