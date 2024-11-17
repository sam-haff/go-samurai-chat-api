echo "do secure copy to the server..."
set /p server_addr= "enter server addr: "
set /p server_user= "enter server user: "
scp -i ssh-key.pem ../images/* %server_user%@%server_addr%:/home/%server_user%/chat-api/images/ 