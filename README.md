# minhthetus-cli

A professional, high-performance CLI tool designed for automation and developer productivity. 

> [!IMPORTANT]
> **We've migrated!** The CLI has transitioned from a Node.js wrapper to a native **Go-compiled binary** for near-zero latency and zero-dependency distribution.

## 🚀 Key Features

- **Native Performance**: Built with Go for "instant" startup times (<10ms).
- **Stunning TUI**: Rich interactive UI powered by the Charmbracelet ecosystem (`Bubble Tea`, `Lipgloss`).
- **N-Level Nesting**: Supports arbitrary nesting of commands (e.g., `minhthetus-cli git checkout`).
- **Professional Help System**: Automatically generated documentation and usage guides for every command.
- **Single Binary**: Zero dependencies—just download and run.
- **Cross-Platform**: Full support for macOS (Intel/M1), Linux, and Windows.

## 📖 Guides

To get started or contribute, please refer to the following documentation:

- [**Migrating to the Go CLI**](guide/MIGRATION_GO.md) - **Required reading for existing users** of the legacy Node.js version.
- [First-Time Local Setup (Go)](guide/GUIDE_START_LOCAL.go.md) / [(Legacy Node)](guide/GUIDE_START_LOCAL.node.md) - Onboarding instructions to get the CLI running on your machine.
- [Adding New CLI Features](guide/GUIDE_NEW_CLI.md) - Best practices for creating new commands.

---

## ⚡ Quick Start

### Option 1: Install via Homebrew (Recommended)
```bash
# Install directly using Homebrew shorthand (automatically taps and installs)
brew install MinhTuLeHoang/tap/minhthetus-cli
```

### Option 2: Install via Go (Compiling directly)

> [!IMPORTANT]
> **Pre-requisite:** Ensure Go's binary directory (`~/go/bin`) is in your shell's `PATH`. If you haven't set this up, run:
> ```bash
> echo 'export PATH=$PATH:~/go/bin' >> ~/.zshrc && source ~/.zshrc
> ```

```bash
# Download and install via Go's toolchain (Go 1.21+ required)
go install github.com/MinhTuLeHoang/minhthetus-cli@latest

# Set up shell autocomplete (required for Go install)
minhthetus-cli setup-completion
```

### Option 3: Manual Installation (From Source clone)
```bash
# Clone the repository
git clone https://github.com/MinhTuLeHoang/minhthetus-cli.git
cd minhthetus-cli

# Build and move binary to /usr/local/bin
make install
```

### Verify Setup
```bash
# Verify it runs and outputs version
minhthetus-cli version

# Tab completion is automatically configured on first execution!
```

## 🛠 Usage Examples

```bash
# Interactive Git checkout (JIRA integrated)
minhthetus-cli git checkout

# Run a version bump for merge requests
minhthetus-cli git merge-request

# Clean up local logs
minhthetus-cli sys clean-log
```
