# ex : shiftwidth=2 tabstop=2 softtabstop=2 :                                      
SHELL := /bin/sh
SRC := $(wildcard *.go)

.PHONY: all
all: lint.out vet.out coverage.out bench.out

.PHONY: bench
bench: bench.out

bench.out: $(SRC)
	go test -bench . | tee bench.out

cover.out: $(SRC)
	go test -v -cover -covermode atomic -coverprofile cover.out ./...

coverage.html: cover.out
	go tool cover -html=cover.out -o coverage.html

coverage.out: cover.out
	go tool cover -func=cover.out | tee coverage.out

.PHONY: clean
clean:
	rm *.out
	go clean -i ./...

.PHONY: fast
fast: vet cov

lint.out: $(SRC)
	golint | tee lint.out

.PHONY: test
test: coverage.out

vet.out: $(SRC)
	go vet -v . | tee vet.out

