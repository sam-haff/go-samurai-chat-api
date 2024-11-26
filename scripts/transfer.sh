#/bin/bash

sshpass -p $SSH_KEY_PWD scp -r -i ssh-key.pem ../images uploaduser@$SERVER_ADDR:/builds_upload/