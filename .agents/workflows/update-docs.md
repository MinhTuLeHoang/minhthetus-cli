---
description: Automatically updates the user wiki pages and developer reference guides based on the latest native Go command implementations.
---

1.  Read the `update-docs` skill instructions in `.agents/skills/update-docs/SKILL.md`.
2.  List all Go command files under `cmd/` and subfolders (e.g., `cmd/git/`, `cmd/web/`, etc.).
3.  List all user wiki pages in `wiki/` and developer guides in `guide/`.
4.  For each subcommand, compare its local Cobra Go implementation with the corresponding flat `wiki/*.md` file.
5.  Apply necessary updates to existing documentation or create new documentation files. Ensure every page includes a "Version History" section using stable release tags.
6.  Synchronize any configuration changes within `~/.minhthetus-cli/` to `guide/folder-structure-config.md`.
7.  Summarize the updates made to the documentation.
