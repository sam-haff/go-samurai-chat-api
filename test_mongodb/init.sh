#!/bin/bash

openssl rand -base64 756 > mongodb-keyfile
sudo chown 999:999 mongodb-keyfile
mv mongodb-keyfile /etc/mongo/mongodb-keyfile

mongod --replSet rs0 --bind_ip 127.0.0.1,mongo --port 27017