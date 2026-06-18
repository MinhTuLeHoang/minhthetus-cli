# Web Project Starter

Detects the environment and starts the development server using the appropriate package manager.

## Usage
```bash
minhthetus-cli web start [options] [-- [args]]
```

## Options

*   `[args]`: Pass additional arguments to the start command (e.g., `--port 3000`).
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Environment Detection**:
    *   Detects the package manager (pnpm, npm, or yarn) in the current directory.
    *   Exits with error if no package manager is detected.
2.  **Execution**:
    *   Runs the start command using the detected package manager:
        *   **pnpm**: `pnpm start [args]`
        *   **npm**: `npm start [args]`
        *   **yarn**: `yarn start [args]`
    *   Passes stdin, stdout, and stderr directly (interactive/streaming output).
3.  **Forwarding Arguments**:
    *   Any arguments provided after `--` are forwarded directly to the underlying package manager's start command.
    *   Exits with error if the start command fails.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
