# Git Tag Dev/Stg

Automatically calculates the next version based on existing stg/qc tags, creates new annotated tags on the CURRENT branch, and pushes to origin.

## Usage
```bash
minhthetus-cli git tag-dev-stg [options]
```

## Options

*   `-P, --patch`: Increment the patch version (e.g. 1.0.0 → 1.0.1).
*   `-N, --minor`: Increment the minor version (e.g. 1.0.0 → 1.1.0). (Default)
*   `-M, --major`: Increment the major version (e.g. 1.0.0 → 2.0.0).
*   `-m, --message <msg>`: Provide a custom tag message. Defaults to `Release v<version>` if omitted.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Version Detection**:
    *   Fetches all tags matching `stg-v*` and `qc-v*` patterns.
    *   Extracts the numeric version from each tag and compares to find the higher "base version".
    *   Displays the latest STG and QC tags found (or "None" if absent).
2.  **Increment Logic**:
    *   Applies the specified increment (Major/Minor/Patch) to the base version to calculate `NEW_VERSION`.
    *   Defaults to minor increment if no flag is provided.
3.  **Metadata**:
    *   If `--message` is not provided, prompts the user for a tag message.
    *   Defaults to `Release v<NEW_VERSION>` if left blank.
4.  **Tagging**:
    *   Displays the current branch.
    *   Creates two annotated tags on the current commit:
        *   `qc-v<NEW_VERSION>`
        *   `stg-v<NEW_VERSION>`
5.  **Synchronization**:
    *   Pushes both tags to `origin` in a single command.
    *   Exits with error if the push fails.
6.  **Reporting**:
    *   Confirms successful creation and push of both tags.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
