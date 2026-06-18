# Git Backup Branch

Creates a backup of the current branch and maintains only 3 recent versions.
Maintains up to 3 versions of backups for the current branch and prompts for cleanup if exceeded.

## Usage
```bash
minhthetus-cli git backup-branch [options]
```

## Options

*   `-l, --list`: List all backup branches for the current branch without creating a new one.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Identify Current Branch**:
    *   Retrieves the name of the current branch using Git commands.
    *   Exits with error if not in a Git repository.
2.  **List Mode (`--list`)**:
    *   Lists all backup branches matching the pattern `backup/<current-branch>-*`.
    *   Exits after listing without creating a backup.
3.  **Generate Backup Name**:
    *   Creates a name in the format `backup/<current-branch>-dd-mm-yyyy-HHh-MM`.
    *   Example: `backup/feat-login-24-04-2026-14h-30`.
4.  **Create Backup**:
    *   Checks if the backup branch already exists.
    *   If it exists, exits with error and advises the user to wait at least 1 minute.
    *   Creates the backup branch locally.
    *   Pushes the new backup branch to `origin` if the remote exists.
5.  **Manage Versions**:
    *   Lists all existing backup branches for the current branch.
    *   Filters branches following the `backup/<current-branch>-dd-mm-yyyy-HHh-MM` pattern.
    *   Sorts backups by date/time (descending).
6.  **Cleanup (If > 3 backups)**:
    *   Identifies backups older than the 3 most recent versions.
    *   Lists the old backups to the user.
    *   Uses a native Go confirmation TUI to ask for confirmation to delete old branches both locally and on origin.
    *   Deletes confirmed backups both locally and remotely.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
