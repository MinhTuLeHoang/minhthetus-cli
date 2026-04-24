# Git Tag Dev/Stg

Automatically calculates the next version based on existing tags, creates new annotated tags on the current branch, and pushes them.

## Usage
```bash
minhthetus-cli git tag-dev-stg [options]
```

## Options
*   `-P, --patch`: Increment the patch version.
*   `-N, --minor`: Increment the minor version (Default).
*   `-M, --major`: Increment the major version.
*   `-m <message>`: Provide a custom tag message.

## Flow

1.  **Version Detection**:
    *   Fetches all tags from `origin`.
    *   Finds the latest `stg-v*` and `qc-v*` tags.
    *   Extracts and compares versions to find the "base version".
2.  **Increment Logic**:
    *   Applies the specified increment (Major/Minor/Patch) to the base version to calculate the `NEW_VERSION`.
3.  **Metadata**:
    *   Prompts for a tag message or uses a default ("Release vX.Y.Z").
4.  **Tagging**:
    *   Creates two annotated tags on the current commit:
        *   `qc-v${NEW_VERSION}`
        *   `stg-v${NEW_VERSION}`
5.  **Synchronization**:
    *   Pushes both tags to `origin`.
6.  **Reporting**:
    *   Confirms successful creation and push of both tags.
