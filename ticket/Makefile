APP := ticket-service

.PHONY: fmt vet test build docker-build run

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -race -cover ./...

build:
	go build -o bin/$(APP) ./cmd

docker-build:
	docker build -t $(APP):local .

run:
	docker compose up --build
