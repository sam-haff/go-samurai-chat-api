#!/bin/bash
echo $REPO_USER

docker build -t $REPO_USER/chat:api --platform linux/amd64 -f ../dockerfiles_prod/api_service.Dockerfile ./../../
docker build -t $REPO_USER/chat:presence --platform linux/amd64 -f ../dockerfiles_prod/presence_service.Dockerfile ./../../
docker build -t $REPO_USER/chat:upload --platform linux/amd64 -f ../dockerfiles_prod/upload_service.Dockerfile ./../../
docker build -t $REPO_USER/chat:ws --platform linux/amd64 -f ../dockerfiles_prod/ws_service.Dockerfile ./../../

cd 

docker image push --all-tags $REPO_USER/chat