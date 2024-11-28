#!/bin/bash

# Script for generating Go test coverage html.
# Starts local MongDB instance(with 1 replica), 
# runs coverage and generates html from results.

cd ../test_mongodb

docker compose up --detach --wait
cd ..
go test -coverprofile="coverage.out" -v ./...
go tool cover -html="coverage.out" -o "coverage.html"
cd ./test_mongodb
docker compose down
cd ../scripts