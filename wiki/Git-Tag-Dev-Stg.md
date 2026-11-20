# Git Tag Dev/Stg

Calculates the next semantic version across repository-specific tag rules (sandbox or production), validates all target tags against both local and remote before creating any tag, and creates/pushes each tag sequentially.

## Usage
```bash
minhthetus-cli git tag-dev-stg [options]
```

## Options

*   `-P, --patch`: Increment the patch version (e.g. `1.0.0` → `1.0.1`).
*   `-N, --minor`: Increment the minor version (e.g. `1.0.0` → `1.1.0`).
*   `-M, --major`: Increment the major version (e.g. `1.0.0` → `2.0.0`).
*   `-m, --message <msg>`: Provide a custom tag message. Defaults to `Release v<version>` if omitted.
*   `-c, --reconfig`: Reconfigure the repository's tag prefix rule. `(Supported from v1.13.0)`
*   `-p, --pick`: Pick a subset of prefixes from the saved rule for this run only. `(Supported from v1.13.0)`
*   `--prod`: Use the repository's production tag rule (`minhthetus.tag.prod`) instead of sandbox (`minhthetus.tag.sandbox`). `(Supported from v1.13.0)`
*   `-h, --help`: Show the help message and exit.

## Configuration:

- path: `~/.minhthetus-cli/git-tag-presets.json` `(Supported from v1.13.0)`
- stored data format:

```json
{
  "sandbox": [
    "qc-v",
    "stg-v",
    "qc-cicd-v"
  ],
  "prod": [
    "release-v",
    "v"
  ]
}
```

Custom prefixes entered during repository configuration are automatically learned and persisted into `~/.minhthetus-cli/git-tag-presets.json` for reuse across projects.

- **Repository-Local Rules (`.git/config`)** `(Supported from v1.13.0)`:
  *   `minhthetus.tag.sandbox`: Comma-separated sandbox tag prefixes for the repository (e.g., `qc-v,stg-v`, `qc-cicd-v`).
  *   `minhthetus.tag.prod`: Comma-separated production tag prefixes for the repository (e.g., `release-v`, `v`).

## Flow

1.  **Repository & Remote Sync**:
    *   Verifies the current directory is a Git repository with at least one commit at `HEAD`.
    *   Detects `origin` and runs `git fetch --tags` `(Supported from v1.13.0)` (or operates in local-only mode with a warning if no `origin` remote exists).
2.  **Rule Resolution & Smart Auto-Discovery** `(Supported from v1.13.0)`:
    *   Reads `minhthetus.tag.sandbox` (default) or `minhthetus.tag.prod` (`--prod`) from `.git/config`.
    *   If unconfigured or `-c/--reconfig` is passed, scans `git tag -l` via `^([a-zA-Z0-9_-]*?v?)\d+\.\d+\.\d+$` and opens an interactive multi-select menu with detected prefixes pre-selected at the top, global presets from `~/.minhthetus-cli/git-tag-presets.json` below, and a `Custom` prefix entry at the bottom (auto-learning new prefixes).
    *   If `-p/--pick` is passed on a multi-prefix rule, allows selecting a subset of prefixes for the current run only without altering `.git/config`.
3.  **Version Detection**:
    *   Queries the latest semantic version tag across **all** prefixes in the active rule using anchored regex matching.
    *   Selects the highest semantic version as the base version (falling back to `0.0.0` if no matching tags exist).
4.  **Increment Logic**:
    *   Applies `-P`/`-N`/`-M` (mutually exclusive), or displays an interactive prompt for `major`, `minor`, `patch` (default), or `custom` semantic version entry `(Supported from v1.13.0)`.
5.  **Metadata**:
    *   Uses `-m/--message` if provided, or prompts interactively (defaulting to `Release v<NEW_VERSION>`).
6.  **Pre-flight Conflict Check** `(Supported from v1.13.0)`:
    *   Checks all target tags across local and remote (local tag refs and `git ls-remote --tags origin`) **before** creating any tag.
    *   If any conflict is detected, blocks creation, displays per-tag status (`Local`, `Remote`, `Local & Remote`, or `Available`), and offers: (1) auto-increment patch `+1` until conflict-free, (2) create & push missing tags only (when at least one tag is missing), or (3) cancel.
7.  **Sequential Tagging & Independent Push** `(Supported from v1.13.0)`:
    *   Iterates over each target tag sequentially — `git tag -a <tag> -m "<msg>"` followed immediately by `git push origin refs/tags/<tag>` — so CI/CD pipelines trigger independently per tag.
    *   Reports detailed per-tag progress and actionable recovery hints if any step fails midway.
8.  **Reporting**:
    *   Confirms successful creation and push of all target tags.

## Version History

* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.13.0`

- **v1.13.0**: Added dynamic per-repository tag rules (`.git/config` + `~/.minhthetus-cli/git-tag-presets.json` with auto-discovery and auto-learning), `--reconfig`/`--pick`/`--prod` flags, `custom` version input, atomic pre-flight local/remote validation, and sequential independent tag creation and push.
- **v1.11.1**: Fixed base version comparison bug where lexical comparison was used instead of semantic versioning (e.g., `0.69.9` vs `0.69.11`).
- **v1.8.5**: Added interactive version bump selection (Major, Minor, Patch) with keyboard selection when no CLI flag is passed.
- **v1.0.0**: Introduced the git `tag-dev-stg` command for automated STG/QC tagging.
