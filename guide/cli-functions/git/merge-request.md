# Git Merge Request

Automatically bumps version, commits changes, and creates a Merge/Pull Request to the `master` branch.

## Usage
```bash
minhthetus-cli git merge-request [options]
```

## Options
*   `-M, --major`: Force major version bump.
*   `-N, --minor`: Force minor version bump.
*   `-P, --patch`: Force patch version bump.
*   `--no-version`: Skip version bump step.
*   `-m <message>`: Provide a commit message.

## Flow

1.  **Preparation**:
    *   Detects the current branch name.
    *   Determines bump type based on branch prefix (e.g., `fix/` -> patch, others -> minor) or manual flags.
    *   Prompts for a commit message if `-m` is provided without text or defaults to `[bump version]`.
2.  **Syncing**:
    *   Fetches latest `origin/master`.
    *   Rebases the current branch onto `origin/master` to ensure a clean merge.
3.  **Title & Description Generation**:
    *   Extracts JIRA ticket number from the branch name if present.
    *   Constructs a descriptive MR title and body (including commit list).
4.  **Version Bumping**:
    *   If `package.json` exists and not skipped, runs `npm version` with the determined increment type.
5.  **Commit & Push**:
    *   Stages all changes.
    *   Commits with the specified message.
    *   Pushes the current branch to `origin`.
6.  **MR/PR Creation**:
    *   **GitHub**: Uses `gh pr create` if the `gh` CLI is available.
    *   **GitLab**: Uses Git push options (`-o mr.create ...`) for automatic MR creation.
    *   **Agit**: Attempts Agit push (`refs/for/master`) as a fallback for Gitea-like systems.
7.  **Fallback**:
    *   Generates and displays a manual URL for opening the MR/PR in the browser if automation fails or as a reference.
