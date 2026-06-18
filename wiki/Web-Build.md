# Web Project Builder

Detects the environment and runs the build script using the appropriate package manager.

## Usage
```bash
minhthetus-cli web build [options] [-- [args]]
```

## Options

*   `[args]`: Pass additional arguments to the build command.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Environment Detection**:
    *   Detects the package manager (pnpm, npm, or yarn) in the current directory.
    *   Exits with error if no package manager is detected.
2.  **Execution**:
    *   Starts the build command using a spinner for progress feedback:
        *   **pnpm**: `pnpm run build [args]`
        *   **npm**: `npm run build [args]`
        *   **yarn**: `yarn run build [args]`
3.  **Completion**:
    *   Reports success or failure with the total elapsed duration.
    *   Exits with error if the build fails.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
