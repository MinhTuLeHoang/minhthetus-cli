---
name: update-docs
description: Updates the CLI technical documentation in 'guide/cli-functions' to stay in sync with shell script changes in 'src/scripts' and 'src/remote-scripts'.
---

# Update CLI Documentation Skill

Use this skill whenever you need to synchronize the documentation with the latest script implementations.

## Instructions

1.  **Survey Source Code**: Analyze all scripts in `src/scripts/` and `src/remote-scripts/`. Pay attention to:
    *   New command line flags or options.
    *   Changes in the execution flow or logic cases.
    *   Descriptions and help information.
    *   **Configuration Logic**: Identify any new or modified logic that reads from or writes to the `~/.minhthetus-cli/` configuration directory.
2.  **Locate Guides**: Documentation is stored in `guide/cli-functions/` following a directory structure that mirrors the scripts (e.g., `git/`, `web/`, `demo/`).
3.  **Cross-Check & Update**:
    *   **Existing Files**: For each script, verify that its corresponding `.md` file accurately reflects the 'Usage', 'Options', and 'Flow' sections.
    *   **New Files**: If a script exists in `src/scripts/` but has no matching `.md` file in `guide/cli-functions/`, or if a script exists in `src/remote-scripts/` but has no matching `.md` file in `guide/cli-functions/remote-scripts/`, create one using the established format.
    *   **Folder Config**: If the scripts' interaction with `~/.minhthetus-cli/` has changed (e.g., new config file, modified JSON structure, or updated shell functions), synchronize these changes in `guide/folder-structure-config.md`.
    *   **Deleted Files**: If a script has been removed, notify the user and suggest deleting the corresponding guide.
4.  **Formatting Rules**:
    *   Use H1 for the command name.
    *   Use H2 for 'Usage', 'Options', and 'Flow'.
    *   Ensure all file paths in documentation are absolute or relative to the project root as appropriate for the context.
5.  **Output**: Provide a summary of the changes made or a diff of the updated documentation files.
