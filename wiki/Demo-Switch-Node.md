# Node Version Switch Demo (Developer Debug Tool)

> [!NOTE]
> This command is intended for development and debugging purposes. It is compiled under the `dev` tag and available via `minhthetus-cli debug switch-node`.

Demonstrates the shell integration mechanism for switching Node.js versions in the parent shell.

## Usage
```bash
minhthetus-cli debug switch-node --node <version>
```

## Options

*   `--node <version>`: The Node.js version to switch to.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Detection**:
    *   Checks for the existence of `MINHTHETUS_SHELL_PIPE`, which is an environment variable set by the shell integration script.
2.  **Instruction**:
    *   If detected, writes the command `nvm use <version>` to the pipe.
3.  **Parent Execution**:
    *   After the CLI process exits, the parent shell function reads from the pipe and executes the buffered commands (e.g., `nvm use`).
    *   Requires that the CLI has been correctly configured with shell integration on the machine.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
