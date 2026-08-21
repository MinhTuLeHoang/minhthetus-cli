# Git Rename Branch

Safely rename the current Git branch both locally and on the remote origin after verifying branch protection rules and uncommitted changes.

## Usage
```bash
minhthetus-cli git rename-branch
```

## Options
*   `-h, --help`: Help for rename-branch.

## Flow
1.  **Repository and Branch Validation**:
    *   Verifies that the current directory is a Git repository and retrieves the current branch name.
    *   Checks if the current branch is protected (`master`, `main`, `dev`, `staging`, `stg`, `production`, `prod`, or starts with `release`/`releases*`). If protected, aborts.
2.  **Working Tree Cleanliness Check**:
    *   Checks for any uncommitted changes. If any exist, prompts the user to commit or stash them and aborts.
3.  **New Branch Input**:
    *   Prompts the user to enter the new name for the branch.
    *   Validates that the target name is not empty, does not match the current name, and does not already exist locally or on the remote origin.
4.  **Rename Execution**:
    *   Renames the local branch to the new name.
    *   If the old branch exists on the remote origin:
        *   Pushes the new branch to origin and sets the upstream tracking.
        *   Deletes the old branch from the remote origin.

## Version History
* **First Stable Version Supported**: `v1.7.0`
* **Latest Stable Version Update**: `v1.8.0`

- **v1.8.0**: Improve branch protection rules, add support for multiple Git providers
- **v1.7.0**: Introduced the rename command.
