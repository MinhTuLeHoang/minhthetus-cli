# Developer Guidelines & AI Safeguards

Welcome to the `minhthetus-cli` codebase! Please adhere to the following development guidelines.

## 🤖 IMPORTANT: AI Agent Compiler Safeguard
> [!IMPORTANT]
> **DO NOT RUN RAW `go build`, `go run`, `go test`, OR `go clean` COMMANDS DIRECTLY!**
> 
> Direct `go` compilation commands are blocked by the developer sandbox and **will prompt the user for permission on every single run**, causing execution blocks.
> 
> **ALWAYS** use the pre-approved `make` targets instead. The user has pre-approved these `make` targets, and they will execute instantly without any prompts:
> 
> *   **Build standard production binary**:
>     ```bash
>     make build
        ```
> *   **Build developer debug binary**:
>     ```bash
>     make build-dev
>     ```
> *   **Build & install globally to `/usr/local/bin`**:
>     ```bash
>     make install
>     ```
> *   **Run and test the compiled local binary**:
>     ```bash
>     ./minhthetus-cli <args>
>     ```
>     *(Permissions are pre-granted for `./minhthetus-cli` execution).*

## Project Architecture
*   **CLI Framework**: Built with `github.com/spf13/cobra`.
*   **Command Definitions**: All subcommands are placed inside `cmd/` (e.g., `cmd/git/`, `cmd/sys/`).
*   **TUI Components**: Standardized confirmation prompts, selections, and inputs are implemented using the Charmbracelet `Bubble Tea` ecosystem. Use the wrappers in `internal/ui/` instead of importing external UI/prompt libraries.
