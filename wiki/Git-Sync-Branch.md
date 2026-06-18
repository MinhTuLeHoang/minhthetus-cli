# Git Branch Sync

Synchronizes the current branch's latest commit to dev and staging branches. Automatically handles rebase if the branch is linear or cherry-picks otherwise.

## Usage
```bash
minhthetus-cli git sync-branch
```

## Options

*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Identify Current State**:
    *   Gets the current branch name (`ORIGINAL_BRANCH`).
    *   Gets the latest commit hash (`LATEST_COMMIT`) and its message.
2.  **Fetch Updates**:
    *   Fetches the latest `origin/dev` and `origin/staging`.
    *   Resolves the hash for both `dev` and `staging` (prefers remote ref, falls back to local).
    *   Exits with error if either hash cannot be resolved.
3.  **Analyse Geometry**:
    *   Checks if `dev` and `staging` are ancestors of the current commit.
    *   Determines if the synchronization can follow a linear path (rebase) or requires cherry-picks.
4.  **Execution Paths**:
    *   **Linear Case (Both are ancestors)**:
        *   Checkout `dev`.
        *   Pull rebase from `origin/dev`.
        *   Rebase `dev` onto current branch, and force push to `origin`.
        *   Checkout `staging`.
        *   Pull rebase from `origin/dev`, and force push to `origin`.
    *   **Case A: Same Node (Cherry-pick + Pull)**:
        *   If `dev` and `staging` are at the same node.
        *   Checkout `dev`, cherry-pick `LATEST_COMMIT`, and push to `origin`.
        *   Checkout `staging`, pull rebase from `origin/dev`, and push to `origin`.
    *   **Case B: Diverged (Dual Cherry-pick)**:
        *   If `dev` and `staging` are at different nodes.
        *   Checkout `dev`, cherry-pick `LATEST_COMMIT`, and push to `origin`.
        *   Checkout `staging`, cherry-pick `LATEST_COMMIT`, and push to `origin`.
5.  **Cleanup**:
    *   Returns to `ORIGINAL_BRANCH`.
    *   Prints final status for both `dev` and `staging`.
    *   Exits with error if any update failed.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
