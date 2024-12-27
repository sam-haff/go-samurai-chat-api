package messages

import "go.mongodb.org/mongo-driver/bson/primitive"

type MessageData struct {
	MsgId          primitive.ObjectID `bson:"_id" json:"msg_id"`
	ToId           string             `bson:"to" json:"to"`
	FromId         string             `bson:"from" json:"from"`
	Text           string             `bson:"msg" json:"msg"`
	CreatedAt      int64              `bson:"created_at" json:"created_at"`
	ConversationID string             `bson:"conv_id" json:"conv_id"`
}
