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
- [First-Time Local Setup](guide/GUIDE_START_LOCAL.md) - Onboarding instructions to get the CLI running on your machine.
- [Adding New CLI Features](guide/GUIDE_NEW_CLI.md) - Best practices for creating new commands.

---

## ⚡ Quick Start

```bash
# Install via Homebrew
brew install MinhTuLeHoang/tap/minhthetus-cli

# Verify installation
minhthetus-cli version

# Shell completion is setup automatically on first run!
# (Or manually via: minhthetus-cli setup-completion)
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
