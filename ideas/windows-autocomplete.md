# Windows Auto-Complete Setup Plan

## 1. Problem Statement
When running `minhthetus-cli setup-completion` (or running any command where `root.go` triggers silent completion setup) on Windows:
```
Unsupported shell: . Please manually setup completion.
```

### Root Cause
1. **Shell Detection**: `cmd/setup_completion.go` reads `os.Getenv("SHELL")`. On Windows (PowerShell/CMD), `$env:SHELL` is not defined.
2. **Missing Shell Handlers**: Only `.zshrc` and `.bashrc` are handled. PowerShell profiles are not supported.
3. **PowerShell Completion Syntax**: PowerShell does not support Unix process substitution (`source <(...)`). Instead, PowerShell uses:
   ```powershell
   minhthetus-cli completion powershell | Out-String | Invoke-Expression
   ```
4. **Uninstaller Inconsistency**: `cmd/uninstall.go` only cleans up `.zshrc` and `.bashrc`, leaving Windows PowerShell profile uncleaned.

---

## 2. Proposed Solution & Architecture

### Step 1: Detect Shell on Windows & Cross-Platform
Instead of just relying on `os.Getenv("SHELL")`:
- Check `runtime.GOOS == "windows"`.
- If on Windows:
  - Check if running inside PowerShell:
    - Environment variable check: `os.Getenv("PSModulePath")` is present in all PowerShell environments (Windows PowerShell 5.1 & PowerShell 7+).
    - Or inspect parent process name (`powershell.exe` vs `pwsh.exe` vs `bash.exe` / Git Bash).
- Add flag support: `minhthetus-cli setup-completion --shell [powershell|bash|zsh|fish]` so user can explicitly specify their shell.

### Step 2: PowerShell Profile Resolution
PowerShell profiles can exist in multiple locations depending on the PowerShell version:
- **PowerShell 7+ (pwsh)**:
  `$HOME\Documents\PowerShell\Microsoft.PowerShell_profile.ps1`
- **Windows PowerShell 5.1**:
  `$HOME\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1`

**Recommended Implementation**:
- Check both profile paths. If neither exists, create directory and profile for the detected version (or prompt/fallback to PowerShell 7+ if installed, else Windows PowerShell).
- Alternatively, run a quick one-liner via `powershell -NoProfile -Command "$PROFILE"` to dynamically query the active user profile path.

### Step 3: PowerShell Completion Script Injection
Format to append into the PowerShell profile:
```powershell
# minhthetus-cli completion
if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {
    minhthetus-cli completion powershell | Out-String | Invoke-Expression
}
```
*Note*: The `Get-Command` guard ensures PowerShell doesn't throw errors if the binary is temporarily moved or uninstalled.

### Step 4: Update `cmd/uninstall.go`
In `cleanupShellConfig()`:
- Add Windows PowerShell profile paths (`Microsoft.PowerShell_profile.ps1`) to `rcFiles`.
- Remove lines containing `minhthetus-cli` upon uninstallation.

---

## 3. Implementation Checklist (For Later)
- [ ] Update `cmd/setup_completion.go`:
  - [ ] Add shell detection logic for Windows (`runtime.GOOS == "windows"` + `PSModulePath` or process inspection).
  - [ ] Add `--shell` flag to `setupCompletionCmd`.
  - [ ] Handle PowerShell profile writing and verification.
- [ ] Update `cmd/uninstall.go`:
  - [ ] Include PowerShell profile paths in `cleanupShellConfig()`.
- [ ] Test on Windows:
  - [ ] Test in Windows PowerShell 5.1.
  - [ ] Test in PowerShell 7 (pwsh).
  - [ ] Test in Git Bash on Windows.
  - [ ] Test `setup-completion`, tab autocompletion, and `uninstall`.
