# Git Rebase Master

Attempts to rebase the current branch onto `origin/HEAD` (the default main branch of the remote `origin` repository). If the rebase is successful without any conflicts, it force pushes the updated branch to remote `origin`. If conflicts occur during rebase, it automatically aborts the rebase and restores the original working branch state.

## Usage
```bash
minhthetus-cli git rebase-master
```

## Options

*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Safety & State Checks**:
    *   Retrieves the current branch name.
    *   Ensures there are no uncommitted changes in the working directory.
2.  **Fetch Latest Remote Target**:
    *   Fetches latest changes from `origin`.
3.  **Resolve Default Branch**:
    *   Resolves `origin/HEAD` dynamically using `git symbolic-ref refs/remotes/origin/HEAD` (running `git remote set-head origin --auto` if not configured locally).
    *   Verifies that the current branch is not the default main branch itself.
4.  **Branch Name Verification**:
    *   If the resolved main branch is neither `master` nor `main`, prompts the user with an interactive confirmation.
    *   If the user cancels, the process exits cleanly without making any changes.
5.  **Execute Rebase**:
    *   Runs `git rebase origin/HEAD`.
6.  **Handling Results**:
    *   **Success (No Conflicts)**:
        *   Prints success notification.
        *   Force pushes the rebased branch to remote origin (`git push origin <current-branch> --force`).
    *   **Conflict Encountered**:
        *   Prints warning/error message.
        *   Executes `git rebase --abort` automatically.
        *   Restores the local branch state without leaving dirty rebase states.

## Version History

* **First Stable Version Supported**: `v1.9.0`
* **Latest Stable Version Update**: `v1.9.0`
