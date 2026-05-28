.PHONY: build install setup-completion help

BINARY_NAME=minhthetus-cli

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build             Build the binary and auto-setup shell completion"
	@echo "  install           Build and install the binary to /usr/local/bin"
	@echo "  setup-completion  Manually setup shell completion"

build:
	@echo "🛠 Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: ./$(BINARY_NAME)"
	@echo "⏳ Auto-setting up shell completion..."
	@./$(BINARY_NAME) setup-completion --silent || true

install: build
	@echo "🚀 Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo mv ./$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed successfully! Run '$(BINARY_NAME) help' to get started."

setup-completion: build
	@./$(BINARY_NAME) setup-completion
