# Go REST API for mobile application
REST API written in Golang for the [chat application](https://github.com/sam-haff/samurai-chat-app).
___
**Stack:**
- Go
- Gin
- Go native tests + Testify
- MongoDB
- Firebase Auth
- FCM
- GitHub Actions
  <br/><br/>
## Getting started...
### Remote server
1. Create **Firebase** account
    - Create **Firebase project**
    - Activate **Authentication** component for the project
    - Active **Cloud Messaging** component for the project
    - Download service account credentials file from **Project settings->Service accounts->Generate new private key**
    - Copy the file to the project root
2. Create **MongoDB Atlas** account
    - Create shared cluster
3. On your remote server, where you intend to host the api,
 perform all the instructions mentioned in **scripts/server_setup_readme.txt**
4. Now, read comments in **scripts/build_and_transfer.sh**. Run the script while complying to the instructions in the script file.
 You can look at the **deploy** job in the .github/workflows/workflow.yml for the example.

### Local server
If you want to run locally, you are still required to perform the first step(unfortunately, no fully local setup for FirebaseAuth and FCM is possible) from the instructions for **Remote server** setup. Then:
1. Launch local mongodb:
~~~
cd test_mongodb
docker compose up
~~~
2. Copy your firebase credentials file to Build the API image:
~~~
cd scripts
MONGODB_CONNECTION_URL=mongodb://127.0.0.1 FIREBASE_CREDS_FILE=firebase-config.json SERVER_BUILD_NUMBER=0 bash ./build_chat_api.sh
cd ..
~~~
3. Run the API server:
~~~
cd images
docker run -p 8080 -d go-chat-app-api
~~~
