package accounts

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
)

func SetupTestingAccounts(accs []TestingAccount) {
	mongoInst, err := database.NewTestMongoDBInstance()

	if err != nil {
		log.Fatalf("Failed to connect to test database with %s", err.Error())
	}
	fmt.Print("Connected to mongodb!")

	ctx := context.Background()

	for _, acc := range accs {
		dbCreateUserRecordsInternal(ctx, mongoInst, acc.UserData)
	}
}

type TestingAccount struct {
	UserData
	Token string
	//Username string
	//Email    string
	//Uid      string
}

func (acc TestingAccount) ToTestingAuthRecord() auth.TestingAuthRecord {
	return auth.TestingAuthRecord{
		Email: acc.UserData.Email,
		Uid:   acc.UserData.Id,
		Token: acc.Token,
	}
}

func NewTestingAccountFromAuthRecord(authRecord auth.TestingAuthRecord, username string, tokens map[string]string) TestingAccount {
	return TestingAccount{
		UserData: UserData{
			Username: username,
			Email:    authRecord.Email,
			Id:       authRecord.Uid,
		},
		Token: authRecord.Token,
	}
}

func GetTestingAccountInfo(pckgPrefix string, index int) TestingAccount {
	indexStr := strconv.Itoa(index)
	authRecord := auth.GetTestingAuthRecord(pckgPrefix, index)

	return NewTestingAccountFromAuthRecord(authRecord, pckgPrefix+"testingacc"+indexStr, map[string]string{"device": pckgPrefix + "fcmtoken" + indexStr})
}
func GetTestingAccountsInfo(pckgPrefix string, startingIndex int, count int) []TestingAccount {
	accs := make([]TestingAccount, count)
	for i := 0; i < count; i++ {
		accs[i] = GetTestingAccountInfo(pckgPrefix, i+startingIndex)
	}
	return accs
}
