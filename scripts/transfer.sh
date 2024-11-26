#/bin/bash

sshpass -p $SSH_KEY_PWD scp -r -i ssh-key.pem ../images uploaduser@$server_addr:/builds_upload/