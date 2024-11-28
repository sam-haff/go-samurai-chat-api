#!/bin/bash

# Build/export the needed images and prepare all other necessary
# utility files.

./build_chat_api.sh
./build_reverse_proxy.sh
cp ../docker-compose.yaml ../images/