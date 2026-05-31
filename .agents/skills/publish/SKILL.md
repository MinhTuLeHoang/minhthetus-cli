---
name: publish
description: Describes the local publishing workflow for creating and pushing stable releases of minhthetus-cli from the master branch.
---

# CLI Release Publishing Skill (Modular Developer-Agent Co-Design)

Use this skill whenever you need to publish a new stable version of `minhthetus-cli` to production. The workflow is modularized to support both **Human Developer** interactive usage and **AI Agent** automated execution.

---

## 👨‍💻 Developer Interactive Workflow
If a human developer is running the release:
1. They switch to the `master` branch locally.
2. They execute:
   ```bash
   make publish
   ```
3. The script will guide them interactively to sync master, select the bump type, gather changelog entries, create and merge the pull request via `gh`, pull the updates back to master, and push the release tag.

---

## 🤖 AI Agent Automated Workflow
If you are an AI agent executing the release programmatically in the background, you can bypass the interactive prompts entirely using our modular sub-tools:

### Step 1: Branch Sync & Verification
Run the pre-approved Makefile target to verify and synchronize the local master branch:
```bash
make sync-master
```

### Step 2: Discover Latest Stable Version
Retrieve the current highest stable tag:
```bash
make get-latest-stable-version
```
*(This outputs only the version number, e.g. `1.0.2`.)*

### Step 3: Analyze Code Changes & Generate Release Notes
1.  **Check Documentation Integrity**: Check if the user wiki pages in `wiki/` and developer guides in `guide/` are fully up-to-date with your Go command implementations. If any command usage, flags, or internal configurations under `cmd/` were modified but not documented, **first run the `update-docs` skill** to synchronize the documentation files before proceeding with the release.
2.  Run a git log check since the latest tag to see what commits are new:
    ```bash
    git log v<latest-version>..HEAD --oneline
    ```
3.  Check file diffs to understand the code changes.
4.  Generate short, clear changelog entries in simple English that cover all main points. Categorize them into:
    *   `Added`
    *   `Changed`
    *   `Fixed`
    *   `Removed`
5.  Calculate the `<new-version>` string based on semantic rules (e.g. bump `patch` or `minor` based on change gravity. Major bumps are blocked).

### Step 4: Execute Non-Interactive Release
Run the publishing script, passing the calculated target version and generated changelog entries as CLI arguments:
```bash
go run scripts/publish/main.go --version <new-version> --added "Added feature X" --changed "Updated library Y" --fixed "Resolved bug Z"
```
*(You can pass multiple `--added`, `--changed`, `--fixed`, or `--removed` flags as arguments).*

The script will automatically execute the checkout, version constant updates, changelog prepending, git commit, push, PR creation via `gh pr create`, auto-merge, local master pull, and tag generation/pushing without pausing for any terminal input.
