# Repository List Manager

Interactively manage and list tracked repositories used for bulk scans.

## Usage
```bash
minhthetus-cli git list-repo
```

## Options

*   `-h, --help`: Show the help message and exit.

## Configuration:

- path: `~/.minhthetus-cli/list-repo.json`
- stored data format:

```json
[
    {
        "name": "",
        "path": "",
        "description": ""
    }
]
```

## Flow

1.  **Load Configuration**:
    *   Reads the tracking list from `~/.minhthetus-cli/list-repo.json`.
2.  **Interactive Selection**:
    *   If no repositories are tracked, shows "Add New" or "Quit" options.
    *   Displays a searchable filter list of tracked repositories (`Name | Path` format).
    *   Includes a "➕ Add New" option to register a new directory.
    *   Selecting "🚪 Quit" or pressing Escape exits.
3.  **Manage Repository**:
    *   Selecting a repository shows its metadata (Name, Description, Path).
    *   Provides options to "Delete" (untrack) or "Back".
    *   Confirms before deleting/untracking the repository.
4.  **Add Repository**:
    *   Prompts for an absolute path to the repository directory.
    *   Validates the directory's existence.
    *   Automatically extracts metadata (Name, Description) from `package.json` or `go.mod` if present.
5.  **Persistence**:
    *   Changes are saved back to the global configuration directory.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
