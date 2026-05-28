# Migration Guide: Moving to the Go CLI (v2.x)

The `minhthetus-cli` is evolving! We are transitioning from a Node.js-based wrapper to a native Go-compiled binary. This move provides significantly faster startup times, a single zero-dependency binary, and a more robust interactive UI.

## 🚀 Why the Change?

| Feature | Legacy (Node.js) | New (Go) |
| :--- | :--- | :--- |
| **Startup Speed** | ~200-500ms | **<10ms** |
| **Dependencies** | Requires Node.js & `gum` | **None** (Self-contained) |
| **Installation** | NPM / PNPM | **Homebrew** / Direct Binary |
| **UI/UX** | External `gum` | Native **Bubble Tea** |

## 🛠 How to Migrate

### 1. Uninstall the Legacy CLI
The most reliable way to remove the Node.js version of the CLI, including all local assets and shell integrations, is to use the built-in `uninstall` command.

```bash
minhthetus-cli uninstall
```

> [!TIP]
> If for some reason the command is already inaccessible, you can manually remove the package (`npm uninstall -g minhthetus-cli`) and the asset folder (`rm -rf ~/.minhthetus-cli`).

### 2. Install the New Go CLI
The new version is distributed via Homebrew for macOS users and direct binaries for others.

```bash
# Using Homebrew (recommended)
brew tap MinhTuLeHoang/minhthetus-cli
brew install minhthetus-cli
```

### 3. Setup Shell Completion
The new CLI uses Go's native completion system. Run the following command to re-initialize completion for your shell:

```bash
minhthetus-cli setup-completion
```

## 🔄 What's Different?

- **Command Syntax**: All core commands (`git`, `sys`, `web`) remain the same.
- **Performance**: You will notice an "instant" feel when running commands.
- **Single Binary**: You no longer need to have `gum` or `node` installed to use the CLI.

## 🆘 Troubleshooting

If you encounter any issues during migration, please reach out or open an issue on the repository. You can verify your current version by running:

```bash
minhthetus-cli version
```
*(Version 2.0.0+ indicates you are on the Go version)*
