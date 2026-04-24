# Web Project Starter

Starts the web project development server with automatic environment detection.

## Usage
```bash
minhthetus-cli web start [options] [-- [args]]
```

## Options
*   `[args]`: Pass additional arguments to the start command (e.g., `--port 3000`).

## Flow

1.  **Environment Detection**:
    *   Detects the package manager (pnpm, npm, or yarn) in the current directory.
2.  **Execution**:
    *   Runs the start command using the detected package manager:
        *   **pnpm**: `pnpm start "$@"`
        *   **npm**: `npm start "$@"`
        *   **yarn**: `yarn start "$@"`
3.  **Forwarding Arguments**:
    *   Any arguments provided to `minhthetus-cli web start` are forwarded directly to the underlying package manager's start command.
