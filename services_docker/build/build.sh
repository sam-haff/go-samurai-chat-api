#!/bin/bash

docker build -t chat-api -f ../dockerfiles/api_service.Dockerfile ./../../
docker build -t chat-presence -f ../dockerfiles/presence_service.Dockerfile ./../../
docker build -t chat-upload -f ../dockerfiles/upload_service.Dockerfile ./../../
docker build -t chat-ws -f ../dockerfiles/ws_service.Dockerfile ./../../