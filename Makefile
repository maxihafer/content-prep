APP_NAME=content-prep
GOLANGCI_LINT_VERSION=1.61.0
LOCALBIN=$(shell pwd)/bin

.PHONY: dep
dep:
	@echo "Installing dependencies..."
	go mod tidy

.PHONY: build
build: dep
	@echo "Building..."
	go build -o dist/$(APP_NAME) main.go

.PHONY: lint
lint: $(LOCALBIN)/golangci-lint dep
	@echo "Running golangci-lint version ${GOLANGCI_LINT_VERSION}..."
	@$(LOCALBIN)/golangci-lint run

$(LOCALBIN)/golangci-lint:
	@echo "Installing golangci-lint version ${GOLANGCI_LINT_VERSION}..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) v${GOLANGCI_LINT_VERSION}