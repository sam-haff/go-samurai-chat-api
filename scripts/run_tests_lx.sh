cd ../test_mongodb

docker compose up --detach --wait
cd ..
go test -v ./...
cd ./test_mongodb
docker compose down
cd ../scripts