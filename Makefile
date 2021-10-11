bootstrap: start run

start:
	docker-compose up

stop:
	docker-compose down

run:
	go run .

format:
	go fmt

build:
	go build -o bin/ .
