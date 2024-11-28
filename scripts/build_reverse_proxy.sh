#!/bin/bash

# Build preconfigured nginx reverse proxy image 
# and export it to the tar for secure transfer.
# Requires:
# - SSl cert files(fullchain.pem, privkey.pem) 
# for your domain in nginx/certs directory. You can get them from
# LetsEncrypt ACME or from any other SSL certificates provider.

if [ ! -d "../images" ]; then
 mkdir ../images 
fi

echo "building go-chat-app-api-reverse-proxy image..."
docker build -t go-chat-app-api-reverse-proxy ../nginx/.
echo "exporting go-chat-app-api-reverse-proxy image to tar"
docker image save go-chat-app-api-reverse-proxy -o ../images/go-chat-app-api-reverse-proxy.tar
