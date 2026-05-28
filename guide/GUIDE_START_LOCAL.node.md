# Legacy Node.js Setup Guide: minhthetus-cli

> [!WARNING]
> This guide is for the **legacy Node.js version** of the CLI (archived on the `master-node-v1` branch). 
> For the active native Go version, please refer to [GUIDE_START_LOCAL.go.md](file:///Users/lap15864-local/temp/minhthetus-cli/guide/GUIDE_START_LOCAL.go.md).

Welcome! Follow these steps to get your local development environment set up for the legacy Node.js version.

---

## 1. Prerequisites
Ensure you have the following installed:
* `nvm` (Node Version Manager)
* `pnpm` (Node Package Manager)

---

## 2. Onboarding Steps

Root into the project directory and run the following commands:

### Step 1: Switch to the correct Node version
The project specifies its Node version in the `.nvmrc` file.
```bash
nvm use
```
*If you do not have that version installed, run `nvm install` first.*

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

---

## 3. Verify Your Setup

Once linked, you can test if the CLI is working correctly:

1. **Check Help**:
   ```bash
   minhthetus-cli help
   ```
   *You should see the colorful help menu.*

2. **Test a Command**:
   ```bash
   minhthetus-cli hello --name "Developer"
   ```

3. **Test Completion**:
   Type `minhthetus-cli h` and press `TAB`. It should autocomplete to `hello`.

---

## 4. Troubleshooting

* **Command not found?** Ensure your `pnpm` global binary path is in your `PATH` environment variable.
* **Completion not working?** Double-check if you have restarted your terminal or sourced your shell profile after running `setup-completion`.
