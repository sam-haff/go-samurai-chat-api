#!/bin/bash

if not exist "/images" mkdir /images

echo "building go-chat-app-api-reverse-proxy image..."
docker build -t go-chat-app-api-reverse-proxy ../nginx/.
echo "exporting go-chat-app-api-reverse-proxy image to tar"
docker image save go-chat-app-api-reverse-proxy -o ../images/go-chat-app-api-reverse-proxy.tar
