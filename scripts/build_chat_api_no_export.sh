#!/bin/bash

# Build the API server image and copy necessary utility files
# to faciliate streamlined deployment update on the server
# side and export it to the the tar for secure transfer.

if [ ! -d "../images" ]; then
 mkdir ../images 
fi

./make_env_file.sh
./make_build_info.sh

echo "building go-chat-app-api image..."
docker build -t go-chat-app-api ../. 


