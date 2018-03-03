default: lint test

.PHONY: lint
lint:
	@golint $(shell go list mcquay.me/pm/...)
	@go vet $(shell go list mcquay.me/pm/...)

.PHONY: test
test:
	@go test -cover $(shell go list mcquay.me/pm/...)
