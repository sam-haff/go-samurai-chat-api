GIT_USER = sam-haff

mkdir /app-certs
mkdir /app-conf
mkdir /app-nginx
mkdir /app-compose

curl -o /app-compose https://raw.githubusercontent.com/$GIT_USER/go-samurai-chat-api/refs/heads/services_local/prod/docker-compose.template
curl -o /app-compose https://raw.githubusercontent.com/$GIT_USER/go-samurai-chat-api/refs/heads/services_local/prod/gen.sh
cd /app-compose
./gen.sh

curl -o /app-nginx https://raw.githubusercontent.com/$GIT_USER/go-samurai-chat-api/refs/heads/nginx/nginx/prod/nginx.conf

curl -o /app-conf https://raw.githubusercontent.com/$GIT_USER/go-samurai-chat-api/refs/heads/scripts/make_env_file.sh
cd /app-conf
./make_env_file.sh







