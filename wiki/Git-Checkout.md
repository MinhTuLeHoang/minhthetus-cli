# Git Checkout

Checks out an existing branch matching a JIRA ID, or creates a new one
following a standardized naming convention. Supports numeric shorthand
(e.g. 9404 -> PCFBANK-9404).

## Usage
```bash
minhthetus-cli git checkout [options]
```

## Options

*   `-j, --jira-ticket <id>`: JIRA ticket ID (e.g. PCFBANK-9404 or 9404). If omitted, the command will prompt for it.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Get JIRA ID**:
    *   Reads from the `--jira-ticket` flag if provided.
    *   If missing, prompts the user via a native interactive input.
    *   Supports numeric shorthand: `9404` is auto-prefixed to `PCFBANK-9404`.
    *   Supports typing "none" or leaving empty to skip JIRA ID integration.
2.  **Search for Existing Branch**:
    *   If a JIRA ID is provided, searches local branches for any match containing the ID.
    *   If exactly one match is found, checks it out immediately and pulls latest updates from origin.
    *   If multiple matches are found, prompts the user to select one using the interactive list selector.
3.  **Create New Branch (If no match)**:
    *   If no matching branch is found or JIRA ID was skipped, enters the creation flow.
4.  **Select Branch Type**:
    *   Prompts the user to select a type: `feature`, `features`, `hotfix`, `test`, `docs`, `improve`, `bugfix`, `refactor` using the list selector.
5.  **Enter Branch Description**:
    *   Prompts the user for a brief description (e.g., "update user profile").
6.  **Formatting & Naming**:
    *   Converts the description to lowercase.
    *   Replaces spaces with dashes and removes special characters.
    *   Constructs the final branch name:
        *   With JIRA: `<type>/<jira-id>-<formatted-description>`
        *   Without JIRA: `<type>/<formatted-description>`
7.  **Branch Source Confirmation**:
    *   If the current branch is not `master` and doesn't start with `releases/`, prompts the user to select the source branch from 3 options:
        *   `yes - create from current branch`: Uses the current active branch as the source.
        *   `master - create from master`: Uses `master` as the source branch.
        *   `no - cancel`: Aborts the creation flow.
    *   Logs the current branch as the source branch during confirmation.
8.  **Final Checkout**:
    *   Creates and checks out the new branch locally from the selected source branch.
    *   Pushes the new branch to `origin`.
    *   Prints a summary confirming the created branch and its source branch.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.4.2`

- **v1.4.2**: Added confirmation prompt with options (current, master, cancel) when creating a new branch from a non-master/non-release branch, and added source branch summary logs.
- **v1.0.0**: Introduced the checkout command.
