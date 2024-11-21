package database

// TODO: remove usernames collection? it was considered a good practive in Firestore but in mongo it doesn't seem as a good idea
const (
	DatabaseName        = "chat"
	UsersCollection     = "users"
	UsernamesCollection = "usernames"
	MessagesCollection  = "messages"
)
