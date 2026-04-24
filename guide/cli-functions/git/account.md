# Git Account Manager

Managed Git identities. Quickly switch between accounts or manage your saved list.

## Usage
```bash
minhthetus-cli git account [options]
```

## Options
*   `-m, --manage`: Enter management mode to list, create, or delete saved accounts.

## Flow

1.  **Configuration Check**:
    *   Ensures `~/.minhthetus-cli/git-accounts.json` exists.
2.  **Identity Detection**:
    *   Checks `local`, `global`, and `system` git configurations to identify the current `user.name` and `user.email`.
3.  **Modes**:
    *   **Selection Mode (Default)**:
        *   Lists saved accounts from configuration.
        *   Prompts user to select an account to apply locally to the current repository.
        *   Includes options to "Add New Account" or "Quit".
    *   **Management Mode (`--manage`)**:
        *   Displays current identity and saved accounts.
        *   Provides options to "Create New" or "Delete" accounts.
        *   **Create New**: Prompts for Title, Name, and Email using `gum input`.
        *   **Delete**: Prompts to select an account to remove from the JSON configuration.
4.  **Application**:
    *   When an account is selected, it executes `git config user.email` and `git config user.name` for the local repository.
