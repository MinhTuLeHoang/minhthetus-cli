# Web Project Builder

Builds the web project with automatic environment detection and progress feedback.

## Usage
```bash
minhthetus-cli web build [options] [-- [args]]
```

## Options
*   `[args]`: Pass additional arguments to the build command.

## Flow

1.  **Environment Detection**:
    *   Detects the package manager (pnpm, npm, or yarn).
2.  **Background Execution**:
    *   Starts the build command in the background:
        *   **pnpm**: `pnpm run build`
        *   **npm**: `npm run build`
        *   **yarn**: `yarn run build`
    *   Redirects output to a temporary log file.
3.  **User Feedback**:
    *   Displays a spinner while the background build process is running.
4.  **Completion**:
    *   Waits for the background process to finish.
    *   Checks the exit code.
    *   Reports success or failure and displays the total duration.
    *   Cleans up the temporary log file.
