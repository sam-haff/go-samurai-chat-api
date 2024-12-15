package accounts

type UserData struct {
	Id       string            `json:"uid" bson:"_id"`
	Username string            `json:"username" bson:"username"`
	Email    string            `json:"email" bson:"email"`
	Img_url  string            `json:"img_url" bson:"img_url"`
	Tokens   map[string]string `json:"tokens" bson:"tokens"`
	Contacts map[string]bool   `json:"contacts" bson:"contacts"`
}

func NewUserData(id string, email string, username string, imgUrl string) UserData {
	return UserData{
		Id:       id,
		Email:    email,
		Username: username,
		Img_url:  imgUrl,
		Tokens:   make(map[string]string),
		Contacts: make(map[string]bool),
	}
}
