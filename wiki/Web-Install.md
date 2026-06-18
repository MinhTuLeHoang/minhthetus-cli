# Web Project Installer

Installs project dependencies with automatic Node.js version switching and package manager detection.

## Usage
```bash
minhthetus-cli web install [options]
```

## Options

*   `-f, --force`: Force install: removes `node_modules` and existing lock files before installing.
*   `--ci`: CI mode: installs dependencies using the frozen lockfile.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Environment Detection**:
    *   Detects the package manager (pnpm, npm, or yarn) by detecting lock files (`pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`).
    *   Exits with error if no package manager is detected.
2.  **Force Mode Handling**:
    *   If `-f` or `--force` is enabled:
        *   Deletes the `node_modules` directory.
        *   Deletes the detected lock file to ensure a completely fresh installation.
3.  **Execution**:
    *   Runs the installation command based on the detected package manager and mode:
        *   **pnpm**: `pnpm install --frozen-lockfile` (CI) or `pnpm i`.
        *   **npm**: `npm ci` (CI) or `npm i`.
        *   **yarn**: `yarn install --frozen-lockfile` (CI) or `yarn install`.
4.  **Reporting**:
    *   Measures elapsed time and reports success or failure with the elapsed duration.
    *   Exits with error if installation fails.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
