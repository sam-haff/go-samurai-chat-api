package auth

import (
	"context"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
)

// for now only fb auth
type Auth interface {
	VerifyToken(ctx context.Context, token string) (*fbauth.Token, error)
	CreateUser(ctx context.Context, user *fbauth.UserToCreate) (*fbauth.UserRecord, error)
}

type FbAuth struct {
	impl *fbauth.Client
}

func NewAuth(fbApp *firebase.App) Auth {
	fbAuth, _ := fbApp.Auth(context.TODO())
	return FbAuth{fbAuth}
}

func (v FbAuth) VerifyToken(ctx context.Context, token string) (*fbauth.Token, error) {
	fbToken, err := v.impl.VerifyIDToken(ctx, token)
	return fbToken, err
}
func (v FbAuth) CreateUser(ctx context.Context, user *fbauth.UserToCreate) (*fbauth.UserRecord, error) {
	rec, err := v.impl.CreateUser(ctx, user)
	return rec, err
}
