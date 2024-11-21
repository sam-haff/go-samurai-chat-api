package messages

import (
	"context"

	"firebase.google.com/go/v4/messaging"
)

type FcmClient interface {
	Send(context.Context, *messaging.Message) (string, error)
}

type FcmClientInstance struct {
	impl *messaging.Client
}

func (fcm *FcmClientInstance) Send(ctx context.Context, msg *messaging.Message) (string, error) {
	return fcm.impl.Send(ctx, msg)
}

//FcmClient
//FcmClientMock
