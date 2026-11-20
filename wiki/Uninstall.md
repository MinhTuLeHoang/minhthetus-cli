# Uninstall Command

Completely removes `minhthetus-cli` shell autocomplete integrations and uninstalls the CLI binary across macOS, Linux, and Windows.

## Usage

```bash
minhthetus-cli uninstall
```

## Options

*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Executable Inspection & Confirmation**:
    *   Resolves the currently running executable path (including symlink resolution for installation method detection). (Supported from v1.12.0)
    *   Prompts for interactive confirmation before making any changes, defaulting focus to **No** to prevent accidental removal. (Supported from v1.12.0)
2.  **Shell Integration Cleanup**:
    *   Scans all supported shell configuration targets:
        *   `~/.zshrc`
        *   `~/.bashrc` and `~/.bash_profile` (preserving unrelated login logic) (Supported from v1.12.0)
        *   `~/.config/fish/config.fish` and deletes `~/.config/fish/completions/minhthetus-cli.fish` (Supported from v1.12.0)
        *   PowerShell 7+ and Windows PowerShell 5.1 profiles (`Documents\PowerShell\Microsoft.PowerShell_profile.ps1`, `Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1`, OneDrive-redirected `Documents`, `$PROFILE`, and `~/.config/powershell/Microsoft.PowerShell_profile.ps1`) (Supported from v1.12.0)
    *   Uses marker-delimited and block-aware removal (brace-balanced for PowerShell `if (...) { ... }`, keyword-balanced for Bash/Zsh `if ... fi` and Fish `if ... end`) so no dangling closing braces or broken syntax remain, while preserving CRLF/LF line endings, UTF-8 BOM, file permissions, and unrelated user scripts. (Supported from v1.12.0)
3.  **Package Manager or Binary Removal**:
    *   **Homebrew**: Runs `brew uninstall minhthetus-cli`.
    *   **WinGet**: Runs `winget uninstall MinhTuLeHoang.minhthetus-cli` (and provides manual guidance if the running binary is locked). (Supported from v1.11.0)
    *   **Go Toolchain / Manual Build**: Deletes the binary directly; on Windows, if the running `.exe` is locked, schedules automatic removal upon process exit for user-writable directories or prints native PowerShell (`Remove-Item -Force`) and CMD (`del /f`) commands if elevated permissions are required. (Supported from v1.12.0)
4.  **Shell Refresh Hint**:
    *   Prints an OS- and shell-aware command or restart instruction (e.g., `. $PROFILE` for PowerShell, `source ~/.bash_profile` for Git Bash, `source ~/.bashrc` for Bash, `source ~/.zshrc` for Zsh, or `exec fish` for Fish). (Supported from v1.12.0)

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.12.0`

- **v1.12.0**: Added block-aware cleanup for PowerShell 5.1/7+, Git Bash (`.bash_profile`), and Fish, symlink-aware install detection, Windows delayed self-delete handling, OS/shell-aware refresh hints, No-default confirmation, and styled UI output.
- **v1.11.0**: Added WinGet detection and `winget uninstall` integration.
- **v1.0.0**: Introduced the native Go `uninstall` command supporting Homebrew, Go, Manual builds, and `.zshrc`/`.bashrc` cleanup.
