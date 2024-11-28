#!/bin/bash

# Load exported all needed images back to Docker service.

docker load --input go-chat-app-api.tar
docker load --input go-chat-app-api-reverse-proxy.tar