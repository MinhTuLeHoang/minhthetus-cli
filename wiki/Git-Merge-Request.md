# Git Merge Request

Automatically bumps version, commits changes, and prepares a Merge Request to the `master` branch.

## Usage
```bash
minhthetus-cli git merge-request [options]
```

## Options

*   `-M, --major`: Force major version bump.
*   `-N, --minor`: Force minor version bump.
*   `-P, --patch`: Force patch version bump.
*   `--no-version`: Skip version bump step.
*   `-m, --message <msg>`: Provide a commit message. If omitted, the command will prompt for it.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Preparation**:
    *   Detects the current branch name.
    *   Determines bump type based on branch prefix (e.g., `fix/`, `hotfix/`, `docs/`, `test/`, `debug/` → patch; others → minor) or manual flags.
    *   Prompts for a commit message if `-m` is omitted. Defaults to `[bump version]` if left blank.
2.  **Syncing**:
    *   Fetches latest `origin/master`.
    *   Rebases the current branch onto `origin/master` to ensure a clean merge.
    *   Aborts and exits if a rebase conflict is detected.
3.  **Title & Description Generation**:
    *   Extracts JIRA ticket number from the branch name if present.
    *   Constructs a descriptive MR title: `Resolve <JIRA-ID> "<Type>/ <description>"`.
    *   Builds the MR body with a `Closes <JIRA-ID>` line and a commit list (`git log origin/master..HEAD`).
4.  **Version Bumping**:
    *   If `--no-version` is not set and a supported version manifest exists, bumps the version.
    *   Prompts user to confirm the version bump (auto-confirms after 3 seconds).
5.  **Commit & Push**:
    *   Stages all changes (`git add .`).
    *   Commits with the specified message (only if there are staged changes).
    *   Pushes the current branch to `origin`.
6.  **MR/PR Link Generation**:
    *   **GitHub**: Generates a pre-filled GitHub PR URL with title and body and prints it.
    *   **GitLab**: Notifies user to use the web interface or GitLab push options.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
