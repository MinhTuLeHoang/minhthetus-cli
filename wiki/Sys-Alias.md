# Sys Alias Command

Interactively view, search, add, edit, and delete shell command aliases, and apply built-in suggestions across your configured shell profiles.

## Usage

```bash
minhthetus-cli sys alias
```

## Options

*   `-h, --help`: Show the help message and exit.

## Configuration:

- path: `~/.minhthetus-cli/aliases.json`
- stored data format:

```json
{
  "aliases": [
    {
      "name": "master",
      "command": "git checkout <origin-HEAD> && git fetch origin && git pull origin <branch>",
      "kind": "preset:master"
    },
    {
      "name": "pull",
      "command": "git pull origin"
    },
    {
      "name": "c",
      "command": "clear"
    }
  ],
  "applied_suggestions": ["master", "pull", "c"]
}
```

## Flow

1.  **Interactive Menu Loop**:
    *   Opens a continuous interactive menu via Bubble Tea (`View / Search Aliases`, `Add Alias`, `Edit Alias`, `Delete Alias`, `Suggestions` when unapplied presets remain, and `Exit`).
2.  **View / Search Aliases**:
    *   Prompts for an optional case-insensitive search query (leave empty to display all configured aliases sorted alphabetically).
    *   Displays aligned columns (`NAME`, `COMMAND`) and target shell profile paths, automatically opening a scrollable viewport for long lists.
3.  **Add Alias**:
    *   Validates the alias name (`^[a-zA-Z_][a-zA-Z0-9_-]*$`, rejecting `minhthetus-cli`) and checks for case-insensitive duplicates with overwrite confirmation.
    *   Prompts for a single-line command, saves `~/.minhthetus-cli/aliases.json`, and syncs shell profiles.
4.  **Edit Alias**:
    *   Selects an existing alias to update its name and/or command (clearing any preset `kind` if the command is customized), then persists and syncs profiles.
5.  **Delete Alias**:
    *   Confirms deletion of the selected alias and removes it from `~/.minhthetus-cli/aliases.json` while keeping `applied_suggestions` sticky.
6.  **Suggestions**:
    *   Displays a multi-select checklist of unapplied built-in presets:
        *   `master`: Checkout the remote `origin/HEAD` default branch (falling back to `main` or `master`), run `git fetch origin`, and run `git pull origin <branch>`.
        *   `pull`: `git pull origin`.
        *   `c`: `clear` (`Clear-Host` on PowerShell).
    *   Selected suggestions are recorded in `applied_suggestions` so they do not reappear in subsequent runs. Once all suggestions have been applied, the `Suggestions` option is hidden from the main menu.
7.  **Cross-Shell Synchronization & Reload Hint**:
    *   Updates an idempotent delimited block (`# >>> minhthetus-cli aliases >>>` ... `# <<< minhthetus-cli aliases <<<`) in detected shell profiles without altering surrounding user configuration:
        *   **macOS / Linux**: `~/.zshrc`, `~/.bashrc`, `~/.bash_profile`.
        *   **Windows**: Windows PowerShell 5.1 (`WindowsPowerShell\Microsoft.PowerShell_profile.ps1`), PowerShell 7+ (`PowerShell\Microsoft.PowerShell_profile.ps1`), and Git Bash (`~/.bashrc`, `~/.bash_profile`), emitting PowerShell 5.1-compatible functions (`Remove-Item Alias:<name>` + `function`, without `&&`/`||`).
    *   On exit after any change, prints a shell-specific reload command (`source ~/.zshrc`, `source ~/.bashrc`, or `. $PROFILE`).

## Version History

* **First Stable Version Supported**: `v1.14.0`
* **Latest Stable Version Update**: `v1.14.0`
