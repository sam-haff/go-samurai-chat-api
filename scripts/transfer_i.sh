#/bin/bash
# Script for transfering the build to the server, interactive.
echo "do secure copy to the server..."
echo "enter server addr"
read server_addr

scp -v -r -i ssh-key.pem ../images/* uploaduser@$server_addr:/builds_upload/new/