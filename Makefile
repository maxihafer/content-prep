APP_NAME=content-prep

.PHONY: build
build:
	@echo "Building..."
	@go build -o bin/$(APP_NAME) main.go