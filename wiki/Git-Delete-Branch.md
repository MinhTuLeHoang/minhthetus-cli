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
        *   It is configured as protected on the remote provider (GitHub or GitLab). The CLI checks the remote URL and queries the provider's API via `gh` or `glab` (if the tools are installed and logged in). If they are not logged in or not found, this check is gracefully bypassed.
    *   If protected, aborts.
2.  **Working Tree Cleanliness Check**:
    *   Retrieves the status of the repository's working tree. If local deletion is selected and there are uncommitted changes, aborts to prevent data loss.
3.  **TUI Target Selection**:
    *   Prompts the user with a multi-select menu to choose deletion targets:
        *   `Local branch` (default selected)
        *   `Remote branch (origin/<branch-name>)` (the cursor starts focused here on startup in `v1.8.3+`)
4.  **Confirmation and Deletion**:
    *   If remote deletion is selected, requests user confirmation.
    *   If remote deletion is confirmed, deletes the remote branch.
    *   If local deletion is selected:
        *   Auto-detects the repository's default branch (e.g., `main`, `master`, or `dev`).
        *   Switches to the default branch.
        *   Directly force-deletes the target branch locally.

## Version History
* **First Stable Version Supported**: `v1.8.0`
* **Latest Stable Version Update**: `v1.8.4`

- **v1.8.4**: Force-delete local branch directly using `-D` to avoid prompt confirmations when the branch is not fully merged.
- **v1.8.3**: Fixed space key toggle behavior in TUI multi-select menu and set default cursor to remote branch on startup.
- **v1.8.0**: Introduced the git `delete-branch` command with smart remote protection and interactive TUI.
