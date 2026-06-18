---
name: update-docs
description: Updates the CLI user documentation in 'wiki/' and developer reference guides in 'guide/' to stay in sync with the native Go command implementations in 'cmd/'.
---

# Update CLI Documentation Skill

**WHEN TO USE:** Only run this skill when trigger from `merge-publish` SKILL or I prompt directly in chat.

**WHAT IS THIS SKILL:** This skill is used to synchronize the documentation with the latest native Go Cobra command implementations.

## Instructions

1.  **Survey Go Codebase**: Analyze command definitions under `cmd/` (e.g. `cmd/git/`, `cmd/web/`, etc.). Pay attention to:
    *   Cobra `Use`, `Short`, `Long` definitions.
    *   Command line flags and arguments registered in `init()` functions.
    *   Interaction flow and logic cases.
    *   **Configuration Logic**: Identify if the command reads/writes to `~/.minhthetus-cli/` configuration directories.
2.  **Locate Target Documentation**:
    *   **User Wiki**: Located under the flat `/wiki/` directory at the project root. Filenames are kebab-case/camel-case and mirror command paths (e.g., `wiki/Git-Account.md`, `wiki/Web-Build.md`).
    *   **Developer Reference**: Internal developer guides stored in `/guide/` (e.g., `guide/GUIDE_NEW_CLI.md`, `guide/folder-structure-config.md`).
3.  **Cross-Check & Update**:
    *   **Existing User Docs**: For each command file in `cmd/`, verify that its corresponding `wiki/Git-<Name>.md` or `wiki/Web-<Name>.md` page matches the usage, option tables, and execution flow.
    *   **Version History Section**: Every subcommand wiki page MUST end with a Version History section using stable releases `vX.Y.Z` (no `-rc` or pre-releases): e.g
        ```markdown
        ## Version History
        * **First Stable Version Supported**: `v1.0.0`
        * **Latest Stable Version Update**: `v1.0.0`
        ```
    *   **New Wiki Pages**: If a command exists under `cmd/` but has no matching flat markdown file in `wiki/`, create one following the standard format, list it in `wiki/_Sidebar.md` and `wiki/Home.md`, and add it.
    *   **Folder Config**: If configuration files in `~/.minhthetus-cli/` have changed, synchronize changes in `guide/folder-structure-config.md`.
    *   **Deleted Commands**: If a command is deprecated or removed from `cmd/`, remove the corresponding `wiki/` page and references in layouts.
4.  **Formatting Rules**:
    *   Use H1 for the command name.
    *   Use H2 for 'Usage', 'Options', 'Flow', and 'Version History'.
    *   Make links to other wiki pages using flat relative paths or bracket notation (e.g., `[[Git-Account]]`).
5.  **Version controll**:
    *   **Section "Flow"**: If has new feature or bugs on existed command, update the flow with supported version
    *   **Version History**: If has new feature on existed command, update "Latest Stable Version Update" and update changelod under "First Stable Version Supported" and "Latest Stable Version Update". Ex:
    ```markdown
    ## Version History

    * **First Stable Version Supported**: `v1.3.0`
    * **Latest Stable Version Update**: `v1.3.2`

    - **v1.3.2**: Added pre-update version checking logic to verify if an update is actually needed before performing the update.
    - **v1.3.0**: Introduced the built-in `update` command and native update automation workflows.
    ```
    If a command is first release, only show like this Ex:
    ```markdown
    ## Version History

    * **First Stable Version Supported**: `v1.3.0`
    * **Latest Stable Version Update**: `v1.3.0`
    ```
6.  **Output**: Read `./examples/Git-Command-A.md` for wiki example, then create/update documentation files.
