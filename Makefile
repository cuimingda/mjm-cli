.PHONY: install test

install:
	go install ./cmd/mjm

test:
	go test ./...
