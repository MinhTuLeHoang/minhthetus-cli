# Git Branch Sync

Synchronizes the current branch's latest commit to `dev` and `staging` branches.

## Usage
```bash
minhthetus-cli git sync-branch
```

## Flow

1.  **Identify Current State**:
    *   Gets the current branch name (`ORIGINAL_BRANCH`).
    *   Gets the latest commit hash (`LATEST_COMMIT`) and its message.
2.  **Fetch Updates**:
    *   Fetches the latest `origin/dev` and `origin/staging`.
3.  **Analyse Geometry**:
    *   Checks if `dev` and `staging` are ancestors of the current commit.
    *   Determines if the synchronization can follow a linear path (rebase) or requires cherry-picks.
4.  **Execution Paths**:
    *   **Case A: Linear (Rebase)**:
        *   If both `dev` and `staging` are ancestors.
        *   Checkout `dev`, pull rebase from origin, rebase `dev` onto current branch, and force push.
        *   Checkout `staging`, pull rebase from `dev`, and force push.
    *   **Case B: Same Node (Cherry-pick + Pull)**:
        *   If `dev` and `staging` are at the same node but not ancestors.
        *   Checkout `dev`, cherry-pick `LATEST_COMMIT`, and push.
        *   Checkout `staging`, pull rebase from `dev`, and push.
    *   **Case C: Diverged (Dual Cherry-pick)**:
        *   Checkout `dev`, cherry-pick `LATEST_COMMIT`, and push.
        *   Checkout `staging`, cherry-pick `LATEST_COMMIT`, and push.
5.  **Cleanup**:
    *   Returns to the `ORIGINAL_BRANCH`.
    *   Reports final status for both branches.
