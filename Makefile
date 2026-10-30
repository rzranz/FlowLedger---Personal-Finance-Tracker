.PHONY: all build build-web run seed test test-cover vet clean

all: build

build-web:
	cd web && npm run build

build: build-web
	go build -ldflags="-s -w" -o bin/flowledger.exe ./cmd/flowledger

run:
	go run ./cmd/flowledger

seed:
	go run ./cmd/flowledger -seed

test:
	go test -v ./...

test-cover:
	go test -cover ./...

vet:
	go vet ./...

clean:
	rm -rf bin/ coverage.out flowledger.db*

