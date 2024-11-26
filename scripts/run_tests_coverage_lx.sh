cd ../test_mongodb

docker compose up --detach --wait
cd ..
go test -coverprofile="coverage.out" -v ./...
go tool cover -html="coverage.out" -o "coverage.html"
cd ./test_mongodb
docker compose down
cd ../scripts