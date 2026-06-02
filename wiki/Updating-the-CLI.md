# Updating the minhthetus-cli

To take advantage of new features, bug fixes, and performance improvements, you should keep your `minhthetus-cli` installation up to date.

Updating the CLI updates the binary executable but **fully preserves** all of your local configurations, personal preferences, and saved profiles stored in `~/.minhthetus-cli/`.

---

## ⚡ The Quick Way: Built-in Update Command

The simplest way to update is by running the built-in `update` command. It will dynamically inspect your system to determine how the CLI was installed and either perform the upgrade automatically or show the precise commands required.

```bash
minhthetus-cli update
```

### What it does:
- **Homebrew**: Automatically runs `brew upgrade minhthetus-cli` for you.
- **Go toolchain**: Automatically runs `go install github.com/MinhTuLeHoang/minhthetus-cli@latest` to compile the newest version.
- **Manual Build**: If you are inside the cloned repository, it will offer to run `git pull && make install`. If you are outside the repository, it will guide you on what to do.

---

## 🛠 Manual Update Methods

If you prefer to perform the update manually or if the built-in command encounters permission issues, use one of the methods below matching your installation style:

### Method 1: Homebrew (Recommended)

If you installed via Homebrew:

```bash
# Update Homebrew formulae databases
brew update

# Upgrade the minhthetus-cli formula
brew upgrade minhthetus-cli
```

### Method 2: Go toolchain

If you installed via `go install`:

```bash
go install github.com/MinhTuLeHoang/minhthetus-cli@latest
```

> [!TIP]
> Ensure your `~/go/bin` directory is in your shell's `PATH` to access the updated binary immediately.

### Method 3: Manual Installation (From Source clone)

If you manually cloned the repository:

```bash
# Navigate to the cloned repository directory
cd /path/to/minhthetus-cli

# Fetch latest commits and check out the stable version
git checkout master
git pull

# Compile and install the new binary to /usr/local/bin
make install
```

---

## 🔒 Configuration Safety Guarantee

All configurations, accounts, and session data reside in:
`~/.minhthetus-cli/`

Updating the binary (regardless of which method you choose) will **NEVER** modify or delete this folder. Only the `uninstall` command will remove these configurations after prompting for your explicit confirmation.

---

## ❓ Troubleshooting

### Permission Denied
If you encounter permission issues during manual installation or standard updates (usually with `/usr/local/bin`), run the command with elevated privileges:

```bash
sudo make install
```

---

## 📅 Version History

- **v1.3.0** (Current): Introduced the built-in `update` command and native update automation workflows.
