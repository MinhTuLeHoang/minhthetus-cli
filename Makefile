.PHONY: build build-dev install setup-completion help

BINARY_NAME=minhthetus-cli

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build             Build the binary and auto-setup shell completion"
	@echo "  build-dev         Build the developer binary with debug commands enabled"
	@echo "  install           Build and install the binary to /usr/local/bin"
	@echo "  setup-completion  Manually setup shell completion"

build:
	@echo "🛠 Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: ./$(BINARY_NAME)"
	@echo "⏳ Auto-setting up shell completion..."
	@./$(BINARY_NAME) setup-completion --silent || true
	@echo "💡 Run './$(BINARY_NAME) <command>' to test locally (e.g. './$(BINARY_NAME) git account')."

build-dev:
	@echo "🛠 Building $(BINARY_NAME) (DEVELOPER/DEBUG BUILD)..."
	go build -tags dev -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: ./$(BINARY_NAME)"
	@echo "⏳ Auto-setting up shell completion..."
	@./$(BINARY_NAME) setup-completion --silent || true
	@echo "💡 Run './$(BINARY_NAME) debug' or other commands to test locally."

install: build
	@echo "🚀 Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo mv ./$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed successfully! Run '$(BINARY_NAME) help' to get started."

setup-completion: build
	@./$(BINARY_NAME) setup-completion
