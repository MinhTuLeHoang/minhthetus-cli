# Setup Completion Command

Automatically detects your active shell and installs dynamic tab completion hooks for `minhthetus-cli` across PowerShell 5.1, PowerShell 7+, Git Bash, Bash, Zsh, and Fish. It also runs silently in the background on CLI startup for near-zero-latency onboarding.

## Usage

```bash
minhthetus-cli setup-completion [options]
```

Examples:
```bash
minhthetus-cli setup-completion
minhthetus-cli setup-completion --shell powershell
minhthetus-cli setup-completion --shell bash
minhthetus-cli setup-completion --silent
```

## Options

*   `-s, --shell <string>`: Explicitly specify the target shell to configure (`powershell`, `bash`, `zsh`, `fish`). (Supported from v1.12.0)
*   `--silent`: Run completion setup silently without printing progress or prompts.
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Shell Detection**:
    *   If `--shell` is provided, validates and normalizes the target shell (`powershell`/`pwsh`/`ps`, `bash`/`git-bash`, `zsh`, `fish`). (Supported from v1.12.0)
    *   Otherwise, inspects the active shell environment:
        *   On Windows, walks ancestor processes in-process (zero subprocess overhead) to distinguish `pwsh.exe` (PowerShell 7+), `powershell.exe` (Windows PowerShell 5.1), `bash.exe` (Git Bash), `zsh.exe`, `fish.exe`, and `cmd.exe`, falling back to `SHELL`, `MSYSTEM` / `MINGW_PREFIX`, `POWERSHELL_DISTRIBUTION_CHANNEL`, `PROMPT`, and `PSModulePath`. (Supported from v1.12.0)
        *   On macOS and Linux, inspects `$SHELL` (`zsh`, `bash`, `fish`, `pwsh`).
    *   In interactive non-silent mode, if the shell cannot be detected (or is `cmd.exe`), prompts you to choose a shell or prints `--shell` usage examples. (Supported from v1.12.0)
2.  **Profile & Hook Installation**:
    *   Automatically creates missing parent directories before writing profile files. (Supported from v1.12.0)
    *   Installs idempotent, marker-delimited guarded completion blocks (`# >>> minhthetus-cli completion >>>` ... `# <<< minhthetus-cli completion <<<`) so startup remains error-free even if the binary is temporarily absent from `PATH`:
        *   **PowerShell (Windows)**: Resolves `Documents` via the Windows Registry (`User Shell Folders\Personal`, supporting OneDrive redirection) and configures `PowerShell\Microsoft.PowerShell_profile.ps1` (PS 7+) and/or `WindowsPowerShell\Microsoft.PowerShell_profile.ps1` (PS 5.1). Silent mode updates only the active edition; explicit `--shell powershell` configures available editions. (Supported from v1.12.0)
        *   **PowerShell (macOS / Linux)**: Configures `~/.config/powershell/Microsoft.PowerShell_profile.ps1`. (Supported from v1.12.0)
        *   **Bash / Git Bash**: Configures `~/.bashrc`. On Windows Git Bash, creates `~/.bash_profile` to source `~/.bashrc` if `.bash_profile`, `.bash_login`, and `.profile` are all missing, or appends the completion block to an existing `~/.bash_profile` that does not load `.bashrc`. (Supported from v1.12.0)
        *   **Zsh**: Configures `~/.zshrc`.
        *   **Fish**: Writes `~/.config/fish/completions/minhthetus-cli.fish` (respecting `$XDG_CONFIG_HOME`). (Supported from v1.12.0)
3.  **Activation**:
    *   Prints shell-specific instructions to reload the configuration (such as `. $PROFILE`, `source ~/.bash_profile`, `source ~/.bashrc`, `source ~/.zshrc`, or restarting your shell). (Supported from v1.12.0)

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.12.0`

- **v1.12.0**: Added `-s, --shell` flag with tab completion, zero-subprocess Windows shell detection (PowerShell 5.1, PowerShell 7+, Git Bash), PowerShell/Git Bash/Fish profile setup, OneDrive `Documents` resolution, marker-delimited guarded hooks, and styled UI output.
- **v1.5.1**: Ensured missing shell RC files are created automatically.
- **v1.0.0**: Introduced the native Go `setup-completion` command supporting Zsh, Bash, and `--silent` auto-setup.
