# Node Version Switch Demo

Demonstrates the shell integration mechanism for switching Node.js versions in the parent shell.

## Usage
```bash
minhthetus-cli demo switch-node --node <version>
```

## How It Works
1.  **Detection**: Checks for the existence of `MINHTHETUS_SHELL_PIPE`, which is an environment variable set by the shell integration script.
2.  **Instruction**: If detected, it writes the command `nvm use <version>` to the pipe.
3.  **Parent Execution**: After the CLI process exits, the parent shell function reads from the pipe and executes the buffered commands (e.g., `nvm use`).
4.  **Requirement**: This requires that the CLI has been correctly set up via `minhthetus-cli setup-completion` and the shell configuration has been sourced.
