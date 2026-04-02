# Local Setup Guide: minhthetus-cli

Welcome! Follow these steps to get your local development environment set up and the CLI tool linked for global use.

## 1. Prerequisites

Make sure you have `nvm` (Node Version Manager) and `pnpm` installed on your system.

## 2. Onboarding Steps

Root into the project directory and run the following commands:

### Step 1: Switch to the correct Node version
The project specifies its Node version in the `.nvmrc` file.
```bash
nvm use
```
*If you don't have that version installed, run `nvm install` first.*

### Step 2: Install dependencies
Use `pnpm` to install the required packages (like `omelette` for tab completion).
```bash
pnpm install
```

### Step 3: Link the CLI globally
To make the `minhthetus-cli` command available anywhere in your terminal, link it globally from the project root:
```bash
pnpm link --global
```

### Step 4: Setup Tab Completion
The CLI comes with professional tab completion support. Install it by running:
```bash
minhthetus-cli setup-completion
```
*Follow the instructions in the output—you will likely need to restart your terminal or source your shell config (e.g., `source ~/.zshrc`).*

---

## 3. Verify Your Setup

Once linked, you can test if the CLI is working correctly:

1. **Check Help**:
   ```bash
   minhthetus-cli help
   ```
   *You should see a list of available commands extracted from the `src/scripts` folder.*

2. **Test a Command**:
   ```bash
   minhthetus-cli hello --name "Your Name"
   ```

3. **Test Completion**:
   Type `minhthetus-cli h` and press `TAB`. It should autocomplete to `hello`.

---

## 4. Troubleshooting

- **Command not found?** Ensure your `pnpm` global binary path is in your `PATH` environment variable.
- **Completion not working?** Double-check if you've restarted your terminal after running `setup-completion`.
