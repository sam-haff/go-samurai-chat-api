#!/bin/bash

# Script for running Go tests.
# Starts local MongDB instance(with 1 replica) then runs tests.
# Doesn't kill the MongoDB instance to enable mongosh checks
# on its contents after the tests.

cd ../test_mongodb

docker-compose up --detach --wait
cd ..
go test -v ./...
cd ./scripts