---
name: new-cli
description: Provides instructions and a standardized workflow for adding new native Go Cobra commands and Bubble Tea features to the minhthetus-cli project.
---

# New CLI Feature Skill

Use this skill when you are asked to add a new command, sub-command, or feature to the native Go implementation of `minhthetus-cli`.

## Instructions

1.  **Understand the Cobra Command Hierarchy**:
    *   Commands map to the file structure under `cmd/`.
    *   Example: `minhthetus-cli git account` -> `cmd/git/account.go` using package `git`.
    *   Cobra commands are registered by calling `.AddCommand()` on parent commands in the module's `init()` function (e.g. in `cmd/git/git.go`).

2.  **Create the Go Command File**:
    *   Identify the appropriate package folder under `cmd/`.
    *   Create a new file with a descriptive snake_case name (e.g. `cmd/git/my_command.go`).
    *   Define the `cobra.Command` structure:
        ```go
        package git

        import (
        	"fmt"
        	"github.com/spf13/cobra"
        )

        var MyCommandCmd = &cobra.Command{
        	Use:   "my-command",
        	Short: "A brief description of your command",
        	Long:  `A longer multi-line description of the command's behavior and flow.`,
        	Run: func(cmd *cobra.Command, args []string) {
        		// Implementation logic
        	},
        }
        ```

3.  **Register the Command**:
    *   Attach the subcommand to its parent in the module's `init()` function (e.g. `cmd/git/git.go` or `cmd/root.go`):
        ```go
        func init() {
        	Cmd.AddCommand(MyCommandCmd)
        }
        ```

4.  **Use UI Helpers & Icons**:
    *   The CLI uses a zero-dependency architecture. For all interactive UI components, use the pre-built Bubble Tea wrapper functions inside the `internal/ui` package:
        *   **Confirmations**: `ui.Confirm(prompt string, timeout time.Duration, defaultVal bool)` (e.g. `internal/ui/confirm.go`).
        *   **Filtering Selectors**: `ui.Choose(prompt string, options []string)` (e.g. `internal/ui/select.go`).
        *   **Text Inputs**: `ui.Input(prompt string, placeholder string)` (e.g. `internal/ui/input.go`).
        *   **Spinners**: Use UI indicators for long-running operations.

5.  **Developer / Debug Commands (Dev Build Only)**:
    *   If a command is intended only for development, diagnostic, or testing purposes:
        *   Place it in the `cmd/debug/` directory.
        *   Add the `//go:build dev` build tag at the very top of each file.
        *   Register it under `cmd/debug/debug.go`.
        *   Build the binary in dev mode via `make build-dev` to access it.

6.  **Test the Implementation**:
    *   **Build the binary**: ALWAYS run `make build` (or `make build-dev`) to compile the latest changes. DO NOT run `go build` directly, as it will trigger a security prompt.
    *   Verify help display: `./minhthetus-cli <command-path> -h`.
    *   Check for help summary: `./minhthetus-cli help` or just `./minhthetus-cli`.
    *   Verify functionality by running the compiled local binary:
        ```bash
        ./minhthetus-cli <parent-command> <my-command> --help
        ```

7.  **Maintain Documentation**:
    *   Create a flat user documentation page in `/wiki/` following the kebab-case naming standard (e.g. `wiki/Git-My-Command.md`).
    *   Add a **Version History** section to the bottom, noting the stable version that introduced the feature.
    *   List the new page in `wiki/_Sidebar.md` and `wiki/Home.md`.
