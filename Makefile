docker build:
	docker build -t prod-service:local .

up:
	docker-compose -f docker-compose.yml up --force-recreate

test:
	go test -v ./tests/integration/banner_test.go
