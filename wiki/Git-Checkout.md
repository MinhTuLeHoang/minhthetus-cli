# Git Checkout (JIRA Integrated)

Checks out an existing branch matching a JIRA ID, or creates a new one following a standardized naming convention using our interactive Go TUI.

## Usage
```bash
minhthetus-cli git checkout [options]
```

## Options

| Option | Description |
| :--- | :--- |
| `-j`, `--jira-ticket` | JIRA ticket ID (e.g. PCFBANK-9404). If omitted, the script will prompt for it. |

## Flow

1.  **Get JIRA ID**:
    *   Reads from the `--jira-ticket` flag if provided.
    *   If missing, prompts the user via a native interactive input.
    *   Supports typing "none" or leaving empty to skip JIRA ID integration.
2.  **Search for Existing Branch**:
    *   If a JIRA ID is provided, searches local branches for any match containing the ID.
    *   If exactly one match is found, checks it out immediately.
    *   If multiple matches are found, prompts the user to select one using our interactive list selector.
3.  **Create New Branch (If no match)**:
    *   If no matching branch is found or JIRA ID was skipped, enters the creation flow.
4.  **Select Branch Type**:
    *   Prompts the user to select a type: `feature`, `hotfix`, `test`, `docs`, `improve`, `bugfix`, `refactor` using the list selector.
5.  **Enter Branch Description**:
    *   Prompts the user for a brief description (e.g., "add login page").
6.  **Formatting & Naming**:
    *   Converts the description to lowercase.
    *   Replaces spaces with dashes and removes special characters.
    *   Constructs the final branch name:
        *   With JIRA: `<type>/<jira-id>-<formatted-description>`
        *   Without JIRA: `<type>/<formatted-description>`
7.  **Final Checkout**:
    *   Executes a clean local checkout of the new branch.

## Example
```bash
# Checkout or create branch for PCFBANK-9404
minhthetus-cli git checkout -j PCFBANK-9404

# Interactive mode (will prompt for JIRA ID)
minhthetus-cli git checkout
```

## Version History
* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
