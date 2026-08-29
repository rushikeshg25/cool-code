BINARY := cool-code
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/rushikeshg25/cool-code/cmd.Version=$(VERSION)

.PHONY: build install test vet sec fmt tidy clean run

build: ## Build the binary
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: ## Install to GOBIN
	go install -trimpath -ldflags "$(LDFLAGS)" .

test: ## Run tests
	go test ./...

vet: ## Run go vet
	go vet ./...

# Excluded rules, in order: unhandled errors (covered by review), subprocess
# and file-path arguments (variable by design; execArgv takes no shell and
# paths go through the workspace jail), file modes (source files the agent
# writes into a project are meant to be 0644/0755), and non-cryptographic
# randomness (used for retry jitter and a status phrase, never for secrets).
GOSEC_EXCLUDE := G104,G204,G301,G302,G304,G306,G404

sec: ## Run gosec static security analysis
	gosec -exclude=$(GOSEC_EXCLUDE) -tests=false -quiet ./...

fmt: ## Format the code
	gofmt -w .

tidy: ## Tidy modules
	go mod tidy

run: build ## Build and run
	./$(BINARY)

clean: ## Remove build artifacts
	rm -f $(BINARY)
	go clean
