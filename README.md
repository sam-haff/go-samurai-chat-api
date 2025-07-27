# Go REST API for the chat web application
API and services for the chat web application. Application is live at https://samuraichat.net (frontend code is in <em>./mini_client</em>). 

<img src="https://github.com/user-attachments/assets/277df0c0-5d3c-4858-8c69-e3de080db830" width=400> 

___

**Stack:**
- Go
- Gin
- Gorilla WebSocket
- Go native tests + Testify
- NATS
- MongoDB
- Firebase Auth | Storage | Cloud Messaging 
- GitHub Actions
  <br/><br/>
## Getting started...
### Remote server
1. Create **Firebase** account
    - Create **Firebase project**
    - Activate **Authentication** component for the project
    - Activate **Cloud Messaging** component for the project
    - Activate **Storage** component for the project
    - Download service account credentials file from **Project settings->Service accounts->Generate new private key**
    - Copy the file to the project root
2. Create **MongoDB Atlas** account
    - Create shared cluster
3. Create Docker Hub account, if you don't have it and authorize it(docker login)
4. Create "chat" repo in Docker Hub
5. Build and publish images
~~~
  cd ./services_docker/build
  REPO_USER=yourdockerhubusername ./build_prod.sh
~~~
6. On your remote, authorize docker hub. Then get the setup_remote.sh and run it with correct envs(see comments in repo/scripts/make_env_file.sh). That will create necessary folder structure and generate docker compose file
~~~
  curl -o /app-compose https://raw.githubusercontent.com/sam-haff/go-samurai-chat-api/refs/heads/scripts/setup_remote.sh
  REPO_USER= NATS_URL= FIREBASE_CREDS_FILE= MONGODB_CONNECT_URL= FIREBASE_STORAGE_BUCKET= ./setup_remote.sh
~~~
7. Place your SSL certificates in your remote /app-certs. 
Modify nginx file in /app-nginx to correspond to your domain.
Make sure .env file in /app-conf is good and place service account json in that folder.
8. You can now run the server with:
~~~
  cd /app-compose
  docker compose up
~~~

### Local server
If you want to run locally, you are still required to perform the first step(unfortunately, no fully local setup for Firebase functionality is possible) from the instructions for **Remote server** setup. Then:
1. Create your .env with repo/scripts/make_env_file.sh while following the instructions in the file. Place file to the root of the project.
~~~
  NATS_URL= FIREBASE_CREDS_FILE= MONGODB_CONNECT_URL= FIREBASE_STORAGE_BUCKET= ./scripts/make_env_file.sh
~~~
2. Make sure your service account json is in the project root. 
3. Build services images
~~~
  cd ./services_docker/build
  ./build.sh
~~~
4. Run services(you may want also uncomment mongodb service section in repo/services_local/docker-compose.yaml if you want to run mongodb locally)
~~~
  cd ./services_local
  docker compose up
~~~
5. You can now access API at http://127.0.0.1:8080

### Web Client
Run server:
~~~
cd mini_client
npm run dev -- --port 80
~~~
You can now access the client on <em>http://localhost:80</em>



