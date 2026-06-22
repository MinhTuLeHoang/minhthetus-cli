# System Clean My Mac Command

Cleans up macOS caches, logs, and developer temporary files using `mac-cleaner-cli`.

## Usage
```bash
minhthetus-cli sys clean-my-mac [options]
```

## Options

*   `-u, --uninstall`: Run the cleanup tool in uninstall mode to clean up configurations.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Environment Check**:
    *   Sources the local NVM setup (`~/.nvm/nvm.sh`) if available to check available Node versions.
2.  **Node.js Version Switch/Install**:
    *   Verifies if the active Node.js version is 20 or higher.
    *   If Node.js is not found or is less than version 20, uses `nvm` to automatically install and switch to the latest available Node.js version.
3.  **Execute Cleaner**:
    *   Invokes `npx mac-cleaner-cli` interactively.
    *   If the `-u` or `--uninstall` flag is supplied, runs `npx mac-cleaner-cli uninstall` instead.

## Version History

* **First Stable Version Supported**: `v1.4.0`
* **Latest Stable Version Update**: `v1.4.0`
