#/bin/bash


# Script for transfering the build to the server.
# Requires:
# - $SERVER_ADDR environment variable, which
# should contain ip address of the server.
# - Server user <uploaduser> created as described in server_setup_readme.txt
# - SSH key <ssh-key.pem> for the <uploaduser>, you get it from /uploadroot/builds_upload/.ssh/id_rsa

ssh-keyscan -H $SERVER_ADDR >> ./known_hosts
scp -v -r -i ssh-key.pem -o UserKnownHostsFile=./known_hosts ../images/* uploaduser@$SERVER_ADDR:/builds_upload/new/