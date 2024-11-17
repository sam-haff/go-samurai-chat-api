package auth

import (
	"context"

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
