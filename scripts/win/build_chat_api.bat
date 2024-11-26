if not exist "/images" mkdir /images

echo "building go-chat-app-api image..."
docker build -t go-chat-app-api ../. 
echo "exporting go-chat-app-api image to tar..."
docker image save go-chat-app-api -o ../images/go-chat-app-api.tar


