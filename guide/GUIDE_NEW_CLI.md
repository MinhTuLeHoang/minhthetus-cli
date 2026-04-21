# Guide: Adding New CLI Features

This guide explains how to add new commands and features to the `minhthetus-cli` tool.

## 1. Structure of a Command

Commands are implemented as shell scripts within the `src/scripts` directory. The CLI automatically maps the command name to the corresponding script.

- A file at `src/scripts/my-command.sh` becomes `minhthetus-cli my-command`.
- A file at `src/scripts/dev/logs.sh` becomes `minhthetus-cli dev logs`.

### Essential: Script Description

Every script **MUST** include a description line. This line is parsed by the `help` command to display information about what your script does. 

Add it near the top of your `.sh` file like this:

```bash
#!/bin/bash
# Description: This is a short summary of what this command does.
```

If this line is missing, the command will still work, but the help menu will show an empty description.

---

## 2. N-Level Subfolder Mapping

The CLI supports arbitrary nesting of subfolders within `src/scripts`. Each folder name becomes a part of the command structure.

### Example: Nested subfolders
If you want to organize commands under multiple categories:

1. Create the directory structure:
   `src/scripts/cloud/aws/ec2/start.sh`

2. Add the description to `start.sh`:
   ```bash
   #!/bin/bash
   # Description: Start an EC2 instance.
   ```

3. Call the command:
   ```bash
   minhthetus-cli cloud aws ec2 start --instance-id i-12345
   ```

### How it Works:
- Arguments are consumed sequentially to navigate folders.
- Once a `.sh` file matching the next argument is found, that script is executed with any remaining arguments.
- Tab completion automatically picks up these subfolders, making it easy for users to discover nested commands.

---

## 3. Step-by-Step: Adding a New Command

1.  **Choose a Category**: Decide where the script should live (directly in `src/scripts` or in a subfolder).
2.  **Create the File**: Create a `.sh` file with a descriptive name (e.g., `src/scripts/my-feature.sh`).
3.  **Add the Required Header**:
    ```bash
    #!/bin/bash
    # Description: My new awesome feature.
    ```
4.  **Implement Help & Logic**: Every command script should support a standardized help message.
    - Define `HELP_*` variables.
    - Source `print-help.sh`.
    - **Note:** You do **NOT** need to source `constants.sh`. It is automatically sourced when you include `print-help.sh`.

    ```bash
    # 1. Define help metadata
    HELP_TITLE="My Feature"
    HELP_USAGE="minhthetus-cli my-feature [options]"
    HELP_DESCRIPTION="Explanatory text about what this does."
    HELP_OPTIONS="--opt | Description"
    HELP_EXAMPLE="minhthetus-cli my-feature --opt"

    # 2. Source the help system (this also provides colors/icons like ${GREEN}, ${CHECK}, etc.)
    source "$(dirname "$0")/../generalScripts/print-help.sh" "$@"

    # 3. Implement your logic
    printf "%b\n" "${GREEN}${CHECK} Success!${NC}"
    ```
5.  **Test**:
    - Run `minhthetus-cli help` to verify the description appears.
    - Run your command: `minhthetus-cli my-feature`.
    - Try tab completion: `minhthetus-cli my<TAB>`.
