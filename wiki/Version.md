# Version Command

Displays the active version of `minhthetus-cli`, the release build date, GitHub release link, and total disk usage metrics (binary size + configuration folder size).

## Usage

```bash
minhthetus-cli version [options]
```

Or via global flags:
```bash
minhthetus-cli --version
minhthetus-cli -v
```

## Options

*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Resolve Version & Build Information**:
    *   Retrieves the version dynamically from Go build metadata (`runtime/debug.ReadBuildInfo`), falling back to build variables defined in `internal/config`.
    *   Retrieves the commit timestamp (`vcs.time`) from Go metadata or fallback build date.
2.  **Calculate Disk Usage**:
    *   Resolves the absolute path of the running executable (resolving symlinks if applicable) and measures the binary file size.
    *   Checks the local configuration directory (`~/.minhthetus-cli/`) and calculates the cumulative size of all configuration files.
    *   Calculates the combined total size (`CLI size + Config folder size`).
3.  **Display Information**:
    *   Prints the CLI version and build date.
    *   Prints the GitHub release URL for the current version.
    *   Prints the total footprint with a detailed breakdown:
        ```text
        Total size: <total> (CLI: <cli-size>, Config: <config-size>)
        ```

## Version History

* **First Stable Version Supported**: `v1.11.3`
* **Latest Stable Version Update**: `v1.11.3`
