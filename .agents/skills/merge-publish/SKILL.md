---
name: merge-publish
description: Describes the local publishing workflow for creating and pushing stable releases of minhthetus-cli from the master branch. (after new features have been merged into master)
---

# CLI Release Publishing Skill (Modular Developer-Agent Co-Design)

Use this skill whenever you need to create Pull Request and publish a new stable version of `minhthetus-cli` to production.

---

## 🤖 Workflow

> [!IMPORTANT]
> **CRITICAL: You MUST execute the release tasks using the defined `make` commands exactly as instructed below. Do NOT run the underlying `go run` or other direct Go shell commands, as environment safeguards and automation hooks are strictly configured via the Makefile.**

### Step 1: Branch Sync & Verification
1. If current branch is `master` -> stop then alert me to checkout to feature branch
2. If current branch has any changes that haven't been commited -> stop then tell me to commit
3. If current branch has any local commits that haven't been push -> auto push then process next Step.

### Step 2: Discover Latest Stable Version
Retrieve the current highest stable tag:
```bash
make get-latest-stable-version
```
*(This outputs only the version number, e.g. `1.0.2`.)*

### Step 3: Analyze Code Changes & Generate Release Notes
1.  Run a git log check since the latest tag to see what commits are new:
    ```bash
    git log v<latest-version>..HEAD --oneline
    ```
2.  Check file diffs to understand the code changes.
3.  Generate short, clear changelog entries in simple English that cover all main points. Categorize them into:
    *   `Added`
    *   `Changed`
    *   `Fixed`
    *   `Removed`
4. Version Bump: **Note: major bumps are blocked**. If user has provide target version or bump type (**patch**, **minor**), calculate the `<new-version>` string accordingly (e.g. `v1.0.3` or `v1.1.0`). Otherwise, if user hasn't provide any, **PAUSE and Ask the User**: Output the current stable version and ask the user whether they want to bump the **minor** or **patch** version.
5. Update document:
- read `update-docs` skill for info
- For workflow or technical chage: update `guide/` only
- For update related to new/update/delete `~/.minhthetus-cli` config file:
  - update `guide/folder-structure-config.md` for feature document
  - update `wiki/<updated-cli>` for changes
- For feature added:
  - update `guide/` for feature document
  - update `wiki/` for user document
- For feature or bug fixed:
  - update `wiki/` for user document
- For bug fixed:
  - update `guide/` for workflow document
5.  **Check Documentation Integrity**: Check if the user wiki pages in `wiki/` and developer guides in `guide/` are fully up-to-date with your Go command implementations. If any command usage, flags, or internal configurations under `cmd/` were modified but not documented, **run the `update-docs` skill** to synchronize the documentation files before release. Update `Latest Stable Version` and note version for new feature in command.

**IMPORTANT:** If has update `wiki/` or `guide/`, request review from me, and wait for me to allow continue.

### Step 4: Execute Non-Interactive Release
Run the publishing script through the Makefile target, passing the calculated target version and generated changelog entries as the `ARGS` variable:
```bash
make publish ARGS="--version <new-version> --added 'Added feature X' --changed 'Updated library Y' --fixed 'Resolved bug Z'"
```
*(You can pass multiple `--added`, `--changed`, `--fixed`, or `--removed` flags inside the ARGS string).*

The script will automatically update the version constants, prepend the changelog, commit these changes on your current feature branch, and push the branch to the remote origin remote.

### Step 5: Create Pull Request and Tag Release
1. **Create the Pull Request**: Create a PR using the `gh` CLI.
   - **Base**: `master`
   - **Head**: the current feature branch
   - **Title**: `Release v<new-version> - <High-level summary of changes>`
   - **Body**: Draft a professional, structured description for a development manager, formatted in markdown:
     ```markdown
     ## 🚀 Release v<new-version>

     ### Summary
     <A concise paragraph summarizing the goal of this release and its business/technical impact.>

     ### 📦 Detailed Changes
     - **Added**:
       - <List new features or additions>
     - **Changed**:
       - <List major modifications or enhancements>
     - **Fixed**:
       - <List resolved bugs or issues>
     
     ### 📖 Documentation Updates
     <Note any updates made to Developer Guides or Wiki pages (if applicable).>
     ```
   - **Assignee**: `@me`
   - **Labels**: Analyze the code changes and add any matching labels (a PR can have multiple labels):
     - `documentation`: If changes affect `.md` files, guides, or wikis.
     - `bug`: If any changes fixed bugs.
     - `agent`: If changes update skills, system rules, subagents, or anything related to the AI agent.
     - `enhancement`: For any new features or improvements.
   *Example command:*
   ```bash
   gh pr create --title "Release v1.0.3 - Add user log parsing & update docs" --body "## 🚀 Release v1.0.3

   ### Summary
   This release adds new user log parsing support to the CLI tool and updates the developer docs.

   ### 📦 Detailed Changes
   - **Added**:
     - Log parsing module under cmd/git/log.go
   - **Changed**:
     - Improved help menu styling
   - **Fixed**:
     - Fixed race conditions during sync

   ### 📖 Documentation Updates
   - Updated GUIDE_NEW_CLI.md with flag autocomplete guidelines." --base master --head <feature-branch> --assignee "@me" --label "documentation" --label "enhancement"
   ```
2. **PAUSE & Wait for Merge**: Ask the user to review and merge the PR. Do not auto-merge.
3. **Checkout Master, Pull & Tag**: Once the user confirms the PR is merged:
   - Switch back to master and pull the updates:
     ```bash
     git checkout master && git pull origin master
     ```
   - Compile local build to verify correctness:
     ```bash
     make build-dev
     ```
   - Tag the new version (including the merged PR URL in the message) and push to remote:
     ```bash
     git tag -a v<new-version> -m "Release v<new-version> - <Merged-PR-URL>"
     git push origin v<new-version>
     ```
   - Wiki documentation is auto-synced to GitHub Wiki via the `sync-wiki` GitHub Actions workflow when a `v*` tag is pushed. No manual step needed.
