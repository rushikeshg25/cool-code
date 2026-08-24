BINARY := cool-code
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/rushikeshg25/cool-code/cmd.Version=$(VERSION)

.PHONY: build install test vet fmt tidy clean run

build: ## Build the binary
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: ## Install to GOBIN
	go install -trimpath -ldflags "$(LDFLAGS)" .

test: ## Run tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format the code
	gofmt -w .

tidy: ## Tidy modules
	go mod tidy

run: build ## Build and run
	./$(BINARY)

clean: ## Remove build artifacts
	rm -f $(BINARY)
	go clean
