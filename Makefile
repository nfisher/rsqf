# ex : shiftwidth=2 tabstop=2 softtabstop=2 :                                      
SHELL := /bin/sh
SRC := $(wildcard *.go)

.PHONY: all
all: vet cov bench

.PHONY: bench
bench: $(SRC)
	go test -bench .

coverage.out: $(SRC)
	go test -v -cover -covermode atomic -coverprofile cover.out ./...

coverage.html: coverage.out
	go tool cover -html=cover.out -o coverage.html

.PHONY: cov
cov: coverage.out
	go tool cover -func=cover.out

.PHONY: clean
clean:
	go clean -i ./...

.PHONY: fast
fast: vet cov

.PHONY: test
test: coverage.out

.PHONY: vet
vet: $(SRC)
	go vet -v .

