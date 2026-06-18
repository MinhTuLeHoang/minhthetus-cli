# Git Account Manager

Managed Git identities. Quickly switch between accounts or manage your saved list.

## Usage
```bash
minhthetus-cli git account [options]
```

## Options

*   `-m, --manage`: Enter management mode to list, create, or delete saved accounts.
*   `-h, --help`: Show the help message and exit.

## Configuration:

- path: `~/.minhthetus-cli/git-accounts.json`
- stored data format:

```json
[
    {
        "title": "",
        "name": "",
        "email": ""
    }
]
```

## Flow

1.  **Identity Detection**:
    *   Checks local and global git configurations to identify the current `user.name` and `user.email`.
2.  **Modes**:
    *   **Selection Mode (Default)**:
        *   Displays the current local and global git identity (`user.name`, `user.email`).
        *   Prompts user to confirm switching account for the repository.
        *   If no accounts are saved, enters the Create New flow directly.
        *   Lists all saved accounts from `~/.minhthetus-cli/git-accounts.json`.
        *   Displays accounts using our custom interactive filter selector.
        *   Includes options to "➕ Add New Account" or "🚪 Quit".
        *   When an account is selected, executes `git config` to set `user.email` and `user.name` for the local repository.
    *   **Management Mode (`--manage`)**:
        *   Displays current identity and saved accounts list.
        *   Provides options to "Create New", "Delete", or "Quit".
        *   **Create New**: Prompts for Title, Name, and Email using native interactive inputs.
        *   **Delete**: Prompts to select an account to remove from the JSON configuration.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
