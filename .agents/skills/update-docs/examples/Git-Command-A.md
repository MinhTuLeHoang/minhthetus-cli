# Git Command A
<!-- 
IMPORTANT: Do not add any comments to real wiki md file. This comment is only for agent instructions.

RULE: if feature was supported from first release, do not add anything like this. else add "(Supported from vX.Y.Z)" after feature/flow/notes description.
-->

<!-- the bellow description is cmd cobra long description -->
Do logic A for git command ...

## Usage
```bash
minhthetus-cli git command-a
```

## Options
<!-- List all option that cmd support, check for cmd cobra Flags, add `-h, --help` as well , `-h, --help` support for all cmd -->

<!-- if flag was supported at first release, do not add anything like this -->
*   `-l, --list`: Enter list mode to list all contents.

<!-- For some flag that lately support add Supported from vX.Y.Z like this. -->
*   `-m, --manage`: Enter management mode to controll all contents. (Supported from v1.11.0)


<!-- START "## Configuration:" section (If this cli has config file) -->
## Configuration:

- path: `~/.minhthetus-cli/git-accounts.json` (Supported from v1.11.0)
- stored data format:

```json
{
    "field-a": 0,
    "field-b": 0
}
```
<!-- END "## Configuration:" section -->


## Flow
<!-- List all flow that this cmd will run, -->

1.  **Read Configuration**:
    <!-- No need to note edge case "ensure config file exist" to this, only show main flow -->
    *   Read `~/.minhthetus-cli/git-accounts.json`.
2.  **Identity Detection**:
    *   Checks local, global, and system git configurations to identify the current `user.name` and `user.email`.
3.  **Update identities**:
    *   Lists saved accounts from configuration using our custom interactive list selector.
    *   Prompts user to select an account to apply locally to the current repository.
    <!-- For lately support feature, add (supported from v1.x.x) note here -->
    *   **Create New**: Prompts for Title, Name, and Email using native interactive inputs. (Support from v1.12.0)

## Version History
* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.15.0`

<!-- Add CHANGELOG here, later version is higher like bellow -->
- **v1.3.2**: Added pre-update version checking logic to verify if an update is actually needed before performing the update.
- **v1.3.0**: Introduced the built-in `update` command and native update automation workflows.