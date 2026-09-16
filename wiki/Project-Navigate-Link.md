# Project Navigate Link

The `minhthetus-cli project navigate-link` command helps manage and quickly navigate to project-related links defined in a local `RELATED_TOOL_LINK.md` file.

## Description
This tool allows you to easily store and access common links such as CI/CD pipelines, staging environments, production logs, or project documentation directly from your terminal. It provides an interactive menu to navigate, add, delete, and set default links for various groups.

## Features
- **Auto-Initialization**: If `RELATED_TOOL_LINK.md` is not present, it will prompt you to create a template file.
- **Navigate**: Grouped view of links. If a group has only one link, it will automatically navigate to it.
- **Add New**: Interactively add a new group or add links to an existing group. New groups automatically set their first link as the default.
- **Delete**: Interactively delete outdated links.
- **Change Default**: Switch the default link for any group.
- **Auto-Browser Detection**: Safely detects your operating system and opens the correct default browser.

## File Format Requirements
You can also edit `RELATED_TOOL_LINK.md` manually following this syntax:

```markdown
# RELATED TOOL LINK MANAGEMENT

<!-- this is version of this md file schema, please don't remove this -->
Version: 1

<description>

## Pipeline
- gitlab ci sandbox: https://...
- gitlab ci staging: https://...
- [default] jenkin prod: https://...

## Git Repo
- gitlab: https://...
```

## Version History
- **v2.x.x**: Introduced `project navigate-link` feature.
