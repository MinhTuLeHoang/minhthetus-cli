# Sys Setup Finger-Print Command

Configures terminal admin/root elevation to prefer fingerprint authentication first and fall back to traditional authentication (password/PIN) on failure or cancellation. This command never prompts for, captures, caches, or stores user passwords; all privileged operations delegate directly to native OS authentication (`sudo` on macOS/Linux or UAC on Windows).

## Usage

```bash
minhthetus-cli sys setup finger-print
```

## Options

*   `-s, --status`: Show current fingerprint authentication status and system diagnostics.
*   `-e, --enable`: Enable fingerprint-first authentication with traditional fallback.
*   `-d, --disable`: Disable fingerprint-first authentication and restore prior configuration.
*   `--verify`: Run a live authentication verification test (`sudo -k && sudo -v` on Unix, `UserConsentVerifier` on Windows).
*   `-y, --yes`: Skip interactive confirmation prompts before applying changes.
*   `-h, --help`: Show usage help for `finger-print`.

## Configuration:

- path: `~/.minhthetus-cli/backups/pam/`
- stored data format:
  - **macOS / Linux**: Timestamped pre-edit PAM backups (`<pam-file>.<YYYYMMDD-HHMMSS>.bak`, mode `0600`) stored outside `/etc/pam.d/` so Linux-PAM/OpenPAM never parses backup files.
  - **Windows**: Revert state snapshot (`windows-fingerprint-state.json`, mode `0600`):

```json
{
  "modified_uac_toggle": true,
  "prev_consent_prompt_admin": 5,
  "modified_wbio_srvc": false,
  "prev_wbio_srvc_start_type": "Manual",
  "modified_native_sudo": false,
  "prev_native_sudo_mode": "disabled"
}
```

## Flow

1.  **Interactive Menu (Default)**:
    *   Displays a Bubble Tea selector (`Check Status & Diagnostics`, `Enable / Configure Fingerprint Auth`, `Disable / Revert Fingerprint Auth`, `Verify Biometric / Admin Auth`, `Exit`).
2.  **Diagnostics & Pre-Flight Checks**:
    *   Inspects OS version, biometric hardware, enrollment state, and active authentication stack.
3.  **Preview & Safe Elevation**:
    *   Shows a colorized diff/command plan and requests explicit confirmation (unless `-y, --yes` is passed).
    *   Saves a backup under `~/.minhthetus-cli/backups/pam/` and applies changes using native OS elevation (`sudo install` on macOS/Linux, UAC `RunAs` on Windows) with direct terminal I/O.
4.  **Verification**:
    *   Re-inspects system state and optionally triggers a live authentication test (`sudo -k && sudo -v` on Unix or `UserConsentVerifier` on Windows).

## Platform Support & Security Boundaries

*   **macOS (`darwin`)**: Uses `/etc/pam.d/sudo_local` on macOS Sonoma (14+) when included by `/etc/pam.d/sudo` (survives OS updates), or `/etc/pam.d/sudo` on earlier releases. Checks both `/usr/lib/pam/pam_tid.so` and `pam_tid.so.2`, integrates `pam_reattach.so` with `optional` + `ignore_ssh` before `pam_tid.so` for `tmux`/`screen`/iTerm2 when installed, and writes files atomically with `0444 root:wheel`.
*   **Linux (`linux`)**: Verifies `pam_fprintd.so` exists on disk before editing PAM (preventing `PAM_MODULE_UNKNOWN` lockouts), checks `fprintd-list` hardware and enrollment (invoking `fprintd-enroll` if 0 fingers are enrolled), inspects included PAM files (`common-auth`, `system-auth`, `fingerprint-auth`) to prevent double prompts, and inserts `auth sufficient pam_fprintd.so` before fallback password auth (`0644 root:root`). Supports rollback for direct `/etc/pam.d/sudo`, `pam-auth-update`, and `authselect`.
*   **Windows (`windows`)**: Uses `CurrentBuild` (not `ProductName`), `winbio.dll` (`WinBioEnumBiometricUnits` with `WINBIO_TYPE_FINGERPRINT`), `PnP` Biometric devices, `WbioSrvc` status, `Policies\Microsoft\Biometrics`, `WinBio\AccountInfo`, UAC policies (`EnableLUA`, `ConsentPromptBehaviorAdmin`, `PromptOnSecureDesktop`), and Windows 11 24H2+ `sudo.exe` (`Build >= 26100`). Can start/configure `WbioSrvc`, enable `sudo config --enable normal`, launch `ms-settings:signinoptions`, toggle UAC `ConsentPromptBehaviorAdmin` (`5` <-> `1`) when fingerprint hardware is present, and run `UserConsentVerifier` verification.
    *   **OS Security Boundary**: Windows has no PAM equivalent for inline terminal elevation; `sudo.exe` and UAC always elevate via `consent.exe` on the Secure Desktop, which cannot be bypassed or re-ordered by a CLI.

## Version History

* **First Stable Version Supported**: `v1.15.0`
* **Latest Stable Version Update**: `v1.15.0`
