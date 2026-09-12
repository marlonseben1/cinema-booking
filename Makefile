.PHONY: build vet test tidy check

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

check: build vet test
