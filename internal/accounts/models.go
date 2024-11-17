package accounts

type UserData struct {
	Id       string            `json:"uid" bson:"_id"`
	Username string            `json:"username" bson:"username"`
	Email    string            `json:"email" bson:"email"`
	Img_url  string            `json:"img_url" bson:"img_url"`
	Tokens   map[string]string `json:"tokens" bson:"tokens"`
}
type UsernameData struct {
	Id     string `json:"_id" bson:"_id"`
	UserID string `json:"user_id"`
}
