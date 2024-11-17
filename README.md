# Go REST API for mobile application
REST API for the [chat application](https://github.com/sam-haff/samurai-chat-app), written in Golang.
___
Stack:
- Go
- Gin
- Go native tests + Testify
- MongoDB
- Firebase Auth
- FCM
- GitHub Actions(WIP, build and deployment automation)


# Getting started...
TODO

Needs .env file specifying following vars: 
- FIREBASE_CREDS_FILE
A path to a json file that provided your project Firebase Admin service account credentials
- MONGODB_CONNECT_URL
MongoDB atlas connection url(with auth information). You get that from MongoDB atlas service. Example(details may differ as db or service introduce changes): mongodb+srv://myusername:mypassword@myclustername.gqbpm.mongodb.net 
