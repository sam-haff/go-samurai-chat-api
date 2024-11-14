# go-samurai-chat-api
A custom backend for my Samurai Chat application, written in Golang. It was at first written to be in 1 to 1 correspondence to the initial implementation of backend(Firebase Functions), but will be rewritten to be in compliance with conventional REST structure.
Relies on Firebase Auth for authorization, uses Firebase Admin SDK to process that and manage authentication(user registration).

Needs .env file specifying following vars: 
- FIREBASE_CREDS_FILE
A path to a json file that provided your project Firebase Admin service account credentials
- MONGODB_CONNECT_URL
MongoDB atlas connection url(with auth information). You get that from MongoDB atlas service. Example(details may differ as db or service introduce changes): mongodb+srv://myusername:mypassword@myclustername.gqbpm.mongodb.net 