# Makefile

.PHONY: build run test migrate clean

build:
	go build -o user-service ./cmd/server/main.go

run: build
	./user-service

test:
	go test ./...

migrate:
	sh ./scripts/migrate.sh

clean:
	go clean
	rm -f user-service