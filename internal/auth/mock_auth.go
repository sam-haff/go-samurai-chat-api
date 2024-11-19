package auth

import (
	"context"
	"errors"
	"log"
	"strconv"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/mock"
)

type MockFbAuth struct {
	mock.Mock
}

func (v *MockFbAuth) VerifyToken(ctx context.Context, token string) (*fbauth.Token, error) {
	args := v.Called(ctx, token)
	t := args.Get(0)
	if t == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fbauth.Token), args.Error(1)
}
func (v *MockFbAuth) CreateUser(ctx context.Context, user *fbauth.UserToCreate) (*fbauth.UserRecord, error) {
	if user == nil {
		log.Fatal("Cant be")
	}
	args := v.Called(ctx, user)
	arg0 := args.Get(0)
	if arg0 == nil {
		return nil, args.Error(1)
	}
	return arg0.(*fbauth.UserRecord), args.Error(1)
}

func (v *MockFbAuth) GetUser(ctx context.Context, uid string) (*fbauth.UserRecord, error) {
	args := v.Called(ctx, uid)

	return args.Get(0).(*fbauth.UserRecord), args.Error(1)
}

type TestingAccount struct {
	Username string
	Email    string
	Token    string
	Uid      string
}

func GetTestingAccountInfo(pckgPrefix string, index int) TestingAccount {
	indexStr := strconv.Itoa(index)
	return TestingAccount{
		Username: pckgPrefix + "testingacc" + indexStr,
		Email:    pckgPrefix + "_testingacc" + indexStr + "@t.com",
		Token:    pckgPrefix + "abcde" + indexStr,
		Uid:      pckgPrefix + "-uid" + indexStr,
	}
}
func GetTestingAccountsInfo(pckgPrefix string, startingIndex int, count int) []TestingAccount {
	accs := make([]TestingAccount, count)
	for i := 0; i < count; i++ {
		accs[i] = GetTestingAccountInfo(pckgPrefix, i+startingIndex)
	}
	return accs
}

const SetupAuthMockTestingAccsCount = 2

func FinalizeSetupAuthMock(authMock *MockFbAuth) {
	authMock.On("VerifyToken", mock.Anything, mock.Anything).Return(nil, errors.New("Error placeholder for invalid creds"))
	authMock.On("GetUser", mock.Anything, mock.Anything).Return(nil, errors.New("Error placeholder for invalid creds"))

}

func SetupAuthMock(pckgPrefix string, accs []TestingAccount, finalizeSetup bool) *MockFbAuth {
	mockAuth := MockFbAuth{}
	for _, acc := range accs {
		authToken := &fbauth.Token{}
		authToken.UID = acc.Uid
		authToken.Firebase.SignInProvider = "email"
		userRecord := &fbauth.UserRecord{UserInfo: &fbauth.UserInfo{}}
		userRecord.Email = acc.Email
		userRecord.EmailVerified = false
		userRecord.UID = acc.Uid
		//userRecord.

		mockAuth.On("VerifyToken", mock.Anything, acc.Token).Return(authToken, nil)
		mockAuth.On("GetUser", mock.Anything, acc.Uid).Return(userRecord, nil)
	}

	mockAuthPtr := &mockAuth
	if finalizeSetup {
		FinalizeSetupAuthMock(mockAuthPtr)
	}

	return mockAuthPtr
}
