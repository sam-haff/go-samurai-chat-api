#!/bin/bash
cd ..

if [ -d "../images" ]; then
 mkdir ../images 
fi

./make_env_file.sh
./make_build_info.sh

echo "building go-chat-app-api image..."
docker build -t go-chat-app-api ../. 
echo "exporting go-chat-app-api image to tar..."
docker image save go-chat-app-api -o ../images/go-chat-app-api.tar


