---
name: build-and-run
description: Instructions for compiling, building, and running the minhthetus-cli Go binary. Use this whenever you need to compile, build, test, or execute the local binary.
---

# Build and Run Skill

Use this skill whenever you need to compile, build, test, or run the local Go binary for `minhthetus-cli`.

## Instructions

1.  **ALWAYS Use `make` for Compiling**:
    *   > [!IMPORTANT]
        > **DO NOT** run `go build` or `go run` directly. The sandboxing controller blocks arbitrary `go` executions to prevent untrusted code execution, which will prompt the user for permission each time.
        > **ALWAYS** use `make` targets to compile the project. The `make` targets are pre-approved and will run seamlessly without prompting the user.

2.  **Available Build Targets**:
    *   To build the production-ready binary:
        ```bash
        make build
        ```
    *   To build the developer/debug binary:
        ```bash
        make build-dev
        ```
    *   To run the automated stable release publishing checklist:
        ```bash
        make publish
        ```
    *   To synchronize local wiki changes live to GitHub Wiki:
        ```bash
        make deploy-wiki
        ```

3.  **Running and Testing the Binary**:
    *   Once compiled via `make`, you can run the compiled binary directly.
    *   Use the `./minhthetus-cli` command to test changes locally:
        ```bash
        ./minhthetus-cli <command>
        ```
    *   Your workspace is configured with permissions to run `./minhthetus-cli` without prompting the user.
