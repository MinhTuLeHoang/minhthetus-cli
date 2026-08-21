# Git Delete Branch

Safely delete the current Git branch locally and/or on the remote origin after checking protections, uncommitted changes, and prompting for targets.

## Usage
```bash
minhthetus-cli git delete-branch
```

## Options
*   `-h, --help`: Help for delete-branch.

## Flow
1.  **Repository and Branch Validation**:
    *   Verifies that the current directory is a Git repository and retrieves the current branch name.
    *   Checks if the current branch is protected. A branch is considered protected if:
        *   It is the default branch of the remote origin (e.g. resolved via `refs/remotes/origin/HEAD`).
        *   It matches any pattern configured under `git config --get minhthetus-cli.protected-branches` (supports glob/wildcard patterns like `release/*`).
    *   If protected, aborts.
2.  **Working Tree Cleanliness Check**:
    *   Retrieves the status of the repository's working tree. If local deletion is selected and there are uncommitted changes, aborts to prevent data loss.
3.  **TUI Target Selection**:
    *   Prompts the user with a multi-select menu to choose deletion targets:
        *   `Local branch` (default selected)
        *   `Remote branch (origin/<branch-name>)`
4.  **Confirmation and Deletion**:
    *   If remote deletion is selected, requests user confirmation.
    *   If remote deletion is confirmed, deletes the remote branch.
    *   If local deletion is selected:
        *   Auto-detects the repository's default branch (e.g., `main`, `master`, or `dev`).
        *   Switches to the default branch.
        *   Deletes the target branch locally. If not fully merged, prompts the user to force delete.

## Version History
* **First Stable Version Supported**: `v1.7.0`
* **Latest Stable Version Update**: `v1.7.0`
