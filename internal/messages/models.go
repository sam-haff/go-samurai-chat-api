package messages

type MessageData struct {
	MsgId          string `bson:"_id" json:"msg_id"`
	ToId           string `bson:"to" json:"to"`
	FromId         string `bson:"from" json:"from"`
	Text           string `bson:"msg" json:"msg"`
	FromUsername   string `bson:"username" json:"username"`
	ImgUrl         string `bson:"img_url" json:"img_url"`
	CreatedAt      int64  `bson:"created_at" json:"created_at"`
	ConversationID string `bson:"conv_id" json:"conv_id"`
}
