# git list-repo

Interactive repository manager for local projects tracked by the CLI.

## Usage

```bash
minhthetus-cli git list-repo
```

## Description

The `list-repo` command provides a terminal-based UI (using `gum`) to manage the list of local repositories stored in `~/.minhthetus-cli/list-repo.json`. These repositories are used by other commands (like `web check-malware`) for bulk processing.

## Options

This command is primarily interactive and does not have specific flags besides the standard help flag.

| Option | Description |
| :--- | :--- |
| `-h, --help` | Show the help message and exit. |

## Flow

1.  **Load Configuration**: Reads the tracking list from `~/.minhthetus-cli/list-repo.json`.
2.  **Interactive Selection**:
    - Displays a searchable list of tracked repositories.
    - Provides a "➕ Add New" option to register a new directory.
3.  **Manage Repository**:
    - Selecting a repository shows its metadata (Name, Description, Path).
    - Allows the user to **Delete** the repository from the tracking list.
4.  **Add Repository**:
    - Prompts for an absolute path.
    - Validates the directory's existence.
    - Automatically extracts metadata from `package.json` if present.
5.  **Persistence**: Changes are saved back to the global configuration directory.

---

## Technical Details

- **Storage**: Managed by [folder-structure-config.md](../../folder-structure-config.md).
- **Dependencies**: 
    - `gum`: For filtering and selection UI.
    - `jq`: For parsing and filtering JSON data.
- **Integration**: Calls the underlying `minhthetus-cli repo-track` and `minhthetus-cli repo-untrack` commands.
