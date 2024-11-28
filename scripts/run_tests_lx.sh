#!/bin/bash

# Script for running Go tests.
# Starts local MongDB instance(with 1 replica) then runs tests.

cd ../test_mongodb

docker compose up --detach --wait
cd ..
go test -v ./...
cd ./test_mongodb
docker compose down
cd ../scripts