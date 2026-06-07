export GOPATH := $(shell go env GOPATH)
export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)

.PHONY: commit-checker
commit-checker:
	@go build -o ./tools/commit-msg ./cmd/committer/main.go
	@rm -f .git/hooks/commit-msg
	@ln -s ../../tools/commit-msg .git/hooks/commit-msg
	@printf "commit-checker has been built and installed to .git/hooks/commit-msg\n"

.PHONY: test
test:
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build
build:
	@go build ./...

.PHONY: help
help:
	@printf "Available targets:\n"
	@printf "\tcommit-checker:\tbuild hook binary and install to .git/hooks\n"
	@printf "\ttest:\t\trun all tests with race detection and coverage\n"
	@printf "\tvet:\t\trun go vet\n"
	@printf "\tbuild:\t\tverify build compiles\n"
