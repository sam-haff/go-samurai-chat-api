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

func (v *MockFbAuth) AddMockUserRecord(uid string, email string, token string) {
	authToken := &fbauth.Token{}
	authToken.UID = uid
	authToken.Firebase.SignInProvider = "EmailAuthProviderID"
	userRecord := &fbauth.UserRecord{UserInfo: &fbauth.UserInfo{}}
	userRecord.Email = email
	userRecord.EmailVerified = false
	userRecord.UID = uid
	userRecord.ProviderID = authToken.Firebase.SignInProvider
	userInfo := &fbauth.UserInfo{}
	userInfo.Email = userRecord.Email
	userInfo.UID = uid
	userInfo.ProviderID = userRecord.ProviderID
	userRecord.ProviderUserInfo = []*fbauth.UserInfo{userInfo}

	v.On("VerifyToken", mock.Anything, token).Return(authToken, nil)
	v.On("GetUser", mock.Anything, uid).Return(userRecord, nil)
}
func (v *MockFbAuth) AddMockTestingAccount(acc TestingAuthRecord) {
	v.AddMockUserRecord(acc.Uid, acc.Email, acc.Token)
}

type TestingAuthRecord struct {
	Email string
	Token string
	Uid   string
}

func GetTestingAuthRecord(pckgPrefix string, index int) TestingAuthRecord {
	indexStr := strconv.Itoa(index)
	return TestingAuthRecord{
		Email: pckgPrefix + "_testingacc" + indexStr + "@t.com",
		Token: pckgPrefix + "abcde" + indexStr,
		Uid:   pckgPrefix + "-uid" + indexStr,
	}
}
func GetTestingAuthRecords(pckgPrefix string, startingIndex int, count int) []TestingAuthRecord {
	accs := make([]TestingAuthRecord, count)
	for i := 0; i < count; i++ {
		accs[i] = GetTestingAuthRecord(pckgPrefix, i+startingIndex)
	}
	return accs
}

const SetupAuthMockTestingAccsCount = 2

func FinalizeSetupAuthMock(authMock *MockFbAuth) {
	authMock.On("VerifyToken", mock.Anything, mock.Anything).Return(nil, errors.New("Error placeholder for invalid creds"))
	authMock.On("GetUser", mock.Anything, mock.Anything).Return(nil, errors.New("Error placeholder for invalid creds"))

}

func SetupAuthMock(pckgPrefix string, accs []TestingAuthRecord, finalizeSetup bool) *MockFbAuth {
	mockAuth := MockFbAuth{}
	for _, acc := range accs {
		mockAuth.AddMockTestingAccount(acc)
	}

	mockAuthPtr := &mockAuth
	if finalizeSetup {
		FinalizeSetupAuthMock(mockAuthPtr)
	}

	return mockAuthPtr
}
