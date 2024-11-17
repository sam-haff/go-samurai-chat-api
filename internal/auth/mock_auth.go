package auth

import (
	"context"
	"errors"

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

type TestingAccount struct {
	Username string
	Email    string
	Token    string
	Uid      string
}

func GetTestingAccountsInfo(pckgPrefix string) []TestingAccount {
	return []TestingAccount{
		{pckgPrefix + "_testingacc1", pckgPrefix + "_testingacc1@t.com", pckgPrefix + "abcde", pckgPrefix + "-uid1"},
		{pckgPrefix + "_testingacc2", pckgPrefix + "_testingacc2@t.com", pckgPrefix + "ghjkl", pckgPrefix + "-uid2"},
	}
}

func SetupAuthMock(pckgPrefix string) *MockFbAuth {
	accs := GetTestingAccountsInfo(pckgPrefix)

	mockAuth := MockFbAuth{}
	for _, acc := range accs {
		authToken := &fbauth.Token{}
		authToken.UID = acc.Uid
		authToken.Firebase.SignInProvider = "email"
		mockAuth.On("VerifyToken", mock.Anything, acc.Token).Return(authToken, nil)
	}

	mockAuth.On("VerifyToken", mock.Anything, mock.Anything).Return(nil, errors.New("Error placeholder for invalid creds"))

	return &mockAuth
}
