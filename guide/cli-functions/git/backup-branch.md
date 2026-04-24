# Git Backup Branch

Creates a backup of the current branch with a date suffix and maintains only the 3 most recent backups.

## Usage
```bash
minhthetus-cli git backup-branch [-l, --list]
```

## Options

| Option | Description |
| :--- | :--- |
| `-l`, `--list` | List all backup branches for the current branch without creating a new one. |

## Flow

1.  **Identify Current Branch**:
    *   Retrieves the name of the current branch using `git rev-parse`.
2.  **Generate Backup Name**:
    *   Creates a name in the format `backup/<current-branch>-dd-mm-yyyy-HHh-MM`.
    *   Example: `backup/feat-login-24-04-2026-14h-30`.
3.  **Create Backup**:
    *   Checks if the backup branch already exists.
    *   If it exists, throws an error and advises the user to wait at least 1 minute.
    *   Executes `git branch <backup-name>` to create the backup point locally.
    *   Pushes the new backup branch to `origin` if the remote exists.
4.  **Manage Versions**:
    *   Lists all existing backup branches for the current branch.
    *   Filters branches following the `backup/<current-branch>-dd-mm-yyyy-HHh-MM` pattern.
    *   Sorts backups by date/time (descending).
5.  **Cleanup (If > 3 backups)**:
    *   Identifies backups older than the 3 most recent versions.
    *   Lists the old backups to the user.
    *   Uses `gum confirm` or a standard prompt to ask for confirmation to delete the old branches both locally and on origin.
    *   Executes `git branch -D` and `git push origin --delete` on the confirmed backups.
