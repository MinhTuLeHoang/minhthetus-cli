# Local Setup Guide: minhthetus-cli

> [!NOTE]
> This project has migrated to a **native Go version**. Both Node.js (Legacy) and Go setup instructions are provided below.

Welcome! Follow these steps to get your local development environment set up.

## 1. Prerequisites

### For Node.js (Legacy)
- `nvm` (Node Version Manager)
- `pnpm`

### For Go (Native)
- `Go 1.21+`

## 2. Onboarding Steps for Node js

Root into the project directory and run the following commands:

### Step 1: Switch to the correct Node version
The project specifies its Node version in the `.nvmrc` file.
```bash
nvm use
```
*If you don't have that version installed, run `nvm install` first.*

### Step 2: Install dependencies
```bash
pnpm install
```

### Step 3: Link the CLI globally
```bash
pnpm link --global
```

### Step 4: Setup Shell Completion
```bash
minhthetus-cli setup-completion
```

Now you can run the tool from anywhere using `minhthetus-cli`.

## 3. Onboarding Steps for Go

The new native version is built using Go and requires no Node.js runtime.

### Step 1: Install Go
Ensure you have **Go 1.21+** installed:
```bash
go version
```
*If not installed, download it from [go.dev](https://go.dev/dl/).*

### Step 2: Build the binary
The project includes a `Makefile` to simplify building and setup. From the root of the project:
```bash
make build
```
*This will compile the binary and **automatically** configure shell completion for you.*

### Step 3: Global Link (Optional)
To use it globally, you can install it using the Makefile:
```bash
make install
```
*Or manually move it:*
```bash
# Example: Move to /usr/local/bin
mv ./minhthetus-cli /usr/local/bin/
```

> [!TIP]
> You no longer need to run `setup-completion` manually. The CLI will automatically detect and configure completion on its first run or during the `make build` process.

---

## 4. Verify Your Setup

Once linked, you can test if the CLI is working correctly:

1. **Check Help**:
   ```bash
   minhthetus-cli help
   ```
   *You should see the colorful tree-based help menu.*

2. **Test a Command**:
   ```bash
   minhthetus-cli hello --name "Developer"
   ```

3. **Test Completion**:
   Type `minhthetus-cli h` and press `TAB`. It should autocomplete to `hello`.

---

## 5. Troubleshooting

- **Command not found?** Ensure your `pnpm` global binary path or `/usr/local/bin` is in your `PATH` environment variable.
- **Completion not working?** Double-check if you've restarted your terminal after running `setup-completion`.
