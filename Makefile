#export GOOSE_DRIVER := postgres
#export GOOSE_DBSTRING := host=localhost port=5432 user=postgres password=postgres dbname=module_13_db sslmode=disable
#export APP_PORT := 8080
#export APP_DSN := $(GOOSE_DBSTRING)

#run:
#	@go run ./cmd/app
#
#psql-run:
#	@docker-compose up -d
#
#migrate:
#	@goose -dir=./migrations up
#
#stress:
#	@gobench -u http://localhost:8080/user/email@example.com -k=true -c 500 -t 10
#
#goose-install:
#	@go get -u github.com/pressly/goose/cmd/goose
#
#bench-install:
#	@go get github.com/valyala/fasthttp
#	@go get github.com/cmpxchg16/gobench


#--------------------------------------------------------------------------------------------------------------

#install:
#	sudo apt install docker-compose \
#	&& sudo usermod -aG docker $$USER \
#	&& sudo service docker restart
#
#
#rm:
#	docker-compose stop \
#	&& docker-compose rm \
#	&& sudo rm -rf pgdata/

docker build:
	docker build -t prod-service:local .

up:
	docker-compose -f docker-compose.yml up --force-recreate

test:
	go test -v ./tests/integration/banner_test.go
