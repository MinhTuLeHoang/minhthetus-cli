# Antigravity Agent Rules & Guidelines

This directory contains persistent instructions that the Antigravity agent automatically loads and adheres to during coding sessions in this repository.

## 1. Automatic Documentation Synchronization

Whenever you perform functional code modifications, updates, or feature additions in this repository (specifically Go files under `cmd/` or related logic/scripts):

*   **Rule**: You **MUST** automatically trigger and run the `/update-docs` workflow (or execute the steps defined in the `update-docs` skill at `.agents/skills/update-docs/SKILL.md`) to verify and synchronize the CLI user documentation in `wiki/` and developer reference guides in `guide/`.
*   **Exceptions (Skip Condition)**: You may skip running `/update-docs` ONLY for minor, non-functional changes such as:
    *   Styling, formatting, or linting modifications (e.g. whitespace, CSS, indentation).
    *   Simple text/commentary adjustments (e.g. typos, minor string text, comments).
    *   Pure renaming changes that do not modify commands, options, execution flow, or CLI configuration directories.

Keep the project's Go command implementations and documentation in perfect alignment!
