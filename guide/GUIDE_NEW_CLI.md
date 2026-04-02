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
3.  **Make it Executable (Optional)**: While the CLI uses `sh` to execute scripts, it's good practice to ensure they have the proper shebang (`#!/bin/bash`).
4.  **Add the Description**:
    ```bash
    # Description: My new awesome feature.
    ```
5.  **Implement Logic**: Write your shell commands. Use `$1`, `$2`, etc., or a loop to parse any arguments passed.
6.  **Test**:
    - Run `minhthetus-cli help` to verify the description appears.
    - Run your command: `minhthetus-cli my-feature`.
    - Try tab completion: `minhthetus-cli my<TAB>`.
