cd ../test_mongodb

docker-compose up --detach
cd ..
go test -v ./...