# minhthetus-cli

A professional, globally accessible Node.js CLI tool that acts as a powerful dispatcher for custom shell scripts. 

## 🚀 Key Features

- **Automated Mapping**: Subfolders and shell scripts within the `src/scripts` directory are automatically mapped to CLI commands.
- **N-Level Nesting**: Supports arbitrary nesting of directories, enabling multi-level commands (e.g., `minhthetus-cli cloud aws start-ec2`).
- **Professional Help System**: Automatically generates usage instructions and extracts command descriptions from script headers.
- **Tab Completion**: Built-in support for shell tab completion using `omelette`.

## 📖 Guides

To get started or contribute, please refer to the following documentation:

- [First-Time Local Setup](guide/GUIDE_START_LOCAL.md) - Onboarding instructions to get the CLI running on your machine.
- [Adding New CLI Features](guide/GUIDE_NEW_CLI.md) - Best practices and instructions for creating new commands and organizing scripts.

---

## Quick Usage Examples

```bash
# List all available commands
minhthetus-cli help

# Execute a greeting script
minhthetus-cli hello --name "Your Name"

# Install tab completion for your shell
minhthetus-cli setup-completion
```
