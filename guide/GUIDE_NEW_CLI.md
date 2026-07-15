# Guide: Adding New CLI Features (Go Version)

This guide explains how to add new commands and features to the native Go implementation of `minhthetus-cli`.

## 1. Native Go Commands (Recommended)

New features should ideally be implemented natively in Go to leverage the performance and type safety of the new architecture.

### Step 1: Generate or Create the command
New commands should be placed in the appropriate module folder under `cmd/` (e.g., `cmd/git/`, `cmd/web/`).

**Example: Adding a new subcommand to `git`**

1. Create a new file `cmd/git/my-subcommand.go`.
2. Use the `git` package name.
3. Export the command variable so it can be registered.

```go
package git

import (
	"fmt"
	"github.com/spf13/cobra"
)

// MySubcommandCmd represents the my-subcommand command
var MySubcommandCmd = &cobra.Command{
	Use:   "my-subcommand",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("my-subcommand called")
	},
}
```

### Step 2: Register the command
Attach the subcommand to its parent in the module's `init()` function (e.g., in `cmd/git/git.go`):

```go
// cmd/git/git.go
func init() {
	Cmd.AddCommand(CheckoutCmd)
	Cmd.AddCommand(MergeRequestCmd)
	Cmd.AddCommand(MySubcommandCmd) // Add your new command here
}
```

After rebuilding, you can run the new command via:
```bash
make build-dev
./minhthetus-cli git my-subcommand
```

### Step 3: Implement logic
Use the Charmbracelet stack for any interactive elements:
- **internal/ui/confirm.go**: For yes/no prompts.
- **internal/ui/select.go**: For filtering/selecting from a list.
- **internal/ui/input.go**: For text input.
- **internal/ui/spinner.go**: For loading states.

### Step 4: Flags and Autocomplete

If your command supports flags or positional arguments, make sure you configure registration and autocompletion properly:

1. **Registering Flags**: Use `Cmd.Flags()` (e.g., `Cmd.Flags().BoolVarP(...)` or `Cmd.Flags().StringVarP(...)`) inside the command's `init()` function to register command-line options.
2. **Autocomplete for Positional Arguments**: Use `ValidArgsFunction` in the `cobra.Command` struct definition to return dynamic autocompletion options for command arguments.

**Example:**
```go
var myFlag bool

var MySubcommandCmd = &cobra.Command{
	Use:   "my-subcommand",
	Short: "A brief description",
	// Used for positional argument completion
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"suggested-arg-1", "suggested-arg-2"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Use myFlag here
	},
}

func init() {
	// Register the flag
	MySubcommandCmd.Flags().BoolVarP(&myFlag, "my-flag", "f", false, "Description of the flag")
}
```

## 2. Designing Interactive UIs

The Go version uses the **Bubble Tea** framework. Instead of calling `gum`, use the pre-built components in the `internal/ui` package.

### Colors and Icons
To maintain visual consistency across all command line interfaces, always use colors, icons, and pre-formatted message helpers from [constants.go](file:///Users/lap15864-local/temp/minhthetus-cli/internal/ui/constants.go).
- Do not define custom color codes or inline emojis/icons inside your command code.
- If you need additional colors or icons, add them directly to [constants.go](file:///Users/lap15864-local/temp/minhthetus-cli/internal/ui/constants.go) first.

**Example usage:**
```go
import "github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"

// Print an error message with the standard error icon and red color
fmt.Println(ui.ErrorMessage("Failed to perform task"))

// Use a style helper with the rocket icon
fmt.Printf("%s %s\n", ui.RocketIcon, ui.GreenStyle().Render("System initialized!"))
```

### Confirmation with Timeout
```go
confirmed, _ := ui.Confirm("Proceed with deploy?", 5*time.Second, true)
```

### Filtering Selection
```go
choice, _ := ui.Choose("Select environment:", []string{"dev", "staging", "prod"})
```

## 3. Developer / Debug Commands (Dev Build Only)

If you are writing commands or features that are intended only for developers (e.g., demos, experiments, tests, or diagnostic tools), you should place them in the `cmd/debug/` directory.

### Step 1: Add the Dev Build Tag
Add the `//go:build dev` build tag at the very top of each file under `cmd/debug/`. This ensures the files are completely ignored in standard production builds.

> [!NOTE]
> Go uses the Build Tags (Build Constraints) directive placed at the top of Go source files.

```go
//go:build dev

package debug
...
```

### Step 2: Build the CLI in Dev Mode
To compile the CLI with these developer commands enabled, run:
```bash
make build-dev
```
This passes the `-tags dev` compilation flag. Once compiled, you can access your dev-only tools using:
```bash
./minhthetus-cli debug
```

In standard production builds (i.e., running `make build`), these developer-only commands will not be compiled, keeping the production binary clean, lightweight, and completely secure.

## 5. Rebuilding

After making changes to the Go code, you must rebuild the binary:

```bash
# Production build (clean binary, no dev commands)
make build

# Developer build (includes dev/debug commands)
make build-dev
```
