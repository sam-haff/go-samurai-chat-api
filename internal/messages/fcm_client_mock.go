package messages

import (
	"context"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
)

type FcmClientMock struct {
	mock.Mock
}

func (fcm *FcmClientMock) Send(ctx context.Context, msg *messaging.Message) (string, error) {
	args := fcm.Called(ctx, msg)

	return args.String(0), args.Error(1)
}
