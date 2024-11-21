cd ../test_mongodb

docker-compose up --detach --wait
cd ..
go test -v ./...
cd ./scripts