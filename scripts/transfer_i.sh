#/bin/bash
#interactive
echo "do secure copy to the server..."
echo "enter server addr"
read server_addr

scp -r -i ssh-key.pem ../images uploaduser@$server_addr:/builds_upload/