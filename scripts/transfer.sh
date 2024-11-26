#/bin/bash

ssh-keyscan -H $SERVER_ADDR >> ./known_hosts
sshpass -p $SSH_KEY_PWD scp -r -i ssh-key.pem -o UserKnownHostsFile=./known_hosts ../images uploaduser@$SERVER_ADDR:/builds_upload/