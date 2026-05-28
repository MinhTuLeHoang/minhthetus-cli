# Local Setup Guide: Native Go minhthetus-cli

> [!NOTE]
> This guide is for the **native Go version** of the CLI (active on the current `feature/migrate-go-cli` branch). 
> For the legacy Node.js version, please refer to [GUIDE_START_LOCAL.node.md](file:///Users/lap15864-local/temp/minhthetus-cli/guide/GUIDE_START_LOCAL.node.md).

Welcome! Follow these steps to set up and run the high-performance Go-compiled version of the developer CLI.

---

## 1. Prerequisites
Ensure you have the following installed:
* **Go 1.25.5** (or compatible Go 1.21+ compiler)
  ```bash
  go version
  ```
  *If not installed, download it from [go.dev](https://go.dev/dl/).*

---

## 2. Onboarding Steps

Root into the project directory and run the following commands:

### Step 1: Build the binary
The project includes a `Makefile` to compile the binary and automatically configure shell completion. From the root of the project:
```bash
make build
```
*This compiles the `minhthetus-cli` binary in the project directory.*

### Step 2: Install globally
To make the binary accessible globally, run:
```bash
make install
```
*This builds the binary and moves it to `/usr/local/bin/minhthetus-cli` (requires `sudo`).*

Alternatively, you can manually move the built binary to any directory within your shell's `PATH`:
```bash
sudo mv ./minhthetus-cli /usr/local/bin/
```

> [!TIP]
> You no longer need to run a manual `setup-completion` command. The Go version will automatically detect your shell (zsh, bash, etc.) and configure autocomplete upon first compilation/run.

---

## 3. Verify Your Setup

Once installed, you can verify your installation:

1. **Check Help**:
   ```bash
   minhthetus-cli help
   ```
   *You should see the colorful tree-based native Go help menu.*

2. **Test a Command**:
   ```bash
   minhthetus-cli hello --name "Developer"
   ```

3. **Test Completion**:
   Type `minhthetus-cli h` and press `TAB`. It should autocomplete to `hello`.

---

## 4. Troubleshooting

* **Command not found?** Ensure `/usr/local/bin` (or the folder where you installed the binary) is in your shell's `PATH`.
* **Completion not working?** Auto-completion scripts are written to your shell's profile (e.g. `~/.zshrc` or `~/.bashrc`). Make sure to restart your terminal or run `source ~/.zshrc` (or the corresponding file for your shell) for the changes to take effect.
