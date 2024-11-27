#/bin/bash

ssh-keyscan -H $SERVER_ADDR >> ./known_hosts
scp -v -r -i ssh-key.pem -o UserKnownHostsFile=./known_hosts ../images/* uploaduser@$SERVER_ADDR:/builds_upload/new/