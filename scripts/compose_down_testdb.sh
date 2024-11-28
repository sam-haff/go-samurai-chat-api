#!/bin/bash

# Kills the local MongoDB instance needed for testing.

cd ../test_mongodb
docker-compose down
cd ../scripts