---
name: new-cli
description: Provides instructions and a standardized workflow for adding new CLI commands and features to the minhthetus-cli project.
---

# New CLI Feature Skill

Use this skill when you are asked to add a new command, sub-command, or feature to the `minhthetus-cli`.

## Instructions

1.  **Understand the Command Hierarchy**:
    *   Commands map to the file structure in `src/scripts/`.
    *   Example: `minhthetus-cli git account` -> `src/scripts/git/account.sh`.
    *   N-level nesting is supported by creating subdirectories.
    *   **Hybrid CLI**: Essential scripts go in `src/scripts/` (bundled). Optional or heavy scripts go in `src/remote-scripts/` (GitHub-only, executed via `npx`).

2.  **Create the Script File**:
    *   Identify the appropriate directory in `src/scripts/`.
    *   Create a `.sh` file with a descriptive name.
    *   Add the mandatory description header on line 2. This must be short and clear (max 2 lines, 1 line preferred):
        ```bash
        #!/bin/bash
        # Description: Brief one-line summary of the command.
        ```

3.  **Implement Standardized Help**:
    *   Define help metadata variables:
        *   `HELP_TITLE`: Title of the script.
        *   `HELP_USAGE`: Example usage string.
        *   `HELP_DESCRIPTION`: Detailed description (multi-line supported).
        *   `HELP_OPTIONS`: Piped list of options (`-f, --flag | Description`). Always list the short flag before the long flag, separated by a comma. Use `\n` for multi-line. **Do NOT** add `-h` or `--help`; it is added automatically.
        *   `HELP_EXAMPLE`: Concrete example command.
        *   `HELP_TAB_SIZE`: (Optional) Change left indentation (default: 3).
    *   Source the help system:
        *   If in root of `src/scripts/`: `source "$(dirname "$0")/generalScripts/print-help.sh" "$@"`
        *   If in 1st level subfolder: `source "$(dirname "$0")/../generalScripts/print-help.sh" "$@"`
        *   If in 2nd level subfolder: `source "$(dirname "$0")/../../generalScripts/print-help.sh" "$@"`
    *   **Note**: `print-help.sh` automatically handles the `-h/--help` flags and sources `constants.sh`.

4.  **Use UI Helpers & Icons**:
    *   Use predefined variables for styling:
        *   Colors: `${GREEN}`, `${YELLOW}`, `${RED}`, `${BLUE}`, `${NC}`.
        *   Icons: `${CHECK}`, `${CROSS}`, `${INFO}`, `${WARN}`, `${BULLET}`.
    *   Always use `printf "%b\n" "..."` when using these variables to ensure correct escape sequence interpretation.

5.  **Test the Implementation**:
    *   Verify help display: `minhthetus-cli <command-path> -h`.
    *   Check for help summary: `minhthetus-cli help` or just `minhthetus-cli`.
    *   Verify functionality by running the command.

6.  **Maintain Documentation**:
    *   Add a technical guide in `guide/cli-functions/` matching the script's path (e.g., `guide/cli-functions/my-module/my-cmd.md`).
    *   This is part of the `update-docs` skill, but essential for a complete feature implementation.

## Remote Commands (Optional)

If a command is heavy (large dependencies) or optional, use the "Remote" approach:
1.  **File Location**: Store it in `src/remote-scripts/...`. These files are excluded from the npm package.
2.  **Registry**: Register it in `src/remote-registry.json`:
    ```json
    "utils/genQR": { "description": "Generate QR code" }
    ```
3.  **Workflow**: When the user runs the command, the CLI will see it's missing locally but present in the registry, then it will call `npx github:MinhTuLeHoang/minhthetus-cli ...` to execute it directly from the source code.
