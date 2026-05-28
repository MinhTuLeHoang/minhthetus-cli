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
go build -o minhthetus-cli main.go
./minhthetus-cli git my-subcommand
```

### Step 3: Implement logic
Use the Charmbracelet stack for any interactive elements:
- **internal/ui/confirm.go**: For yes/no prompts.
- **internal/ui/select.go**: For filtering/selecting from a list.
- **internal/ui/input.go**: For text input.
- **internal/ui/spinner.go**: For loading states.

## 2. Legacy Script Bridge (Fast Migration)

If you have an existing shell script that is too complex to port immediately, you can "bridge" it by creating a Go command that executes the script.

### Example: Bridging a script
1. Create a new command: `cobra-cli add my-feature`.
2. Update `cmd/my_feature.go`:

```go
var myFeatureCmd = &cobra.Command{
    Use:   "my-feature",
    Short: "Bridge to legacy script",
    Run: func(cmd *cobra.Command, args []string) {
        scriptPath := filepath.Join("src", "scripts", "my-feature.sh")
        execCmd := exec.Command("bash", scriptPath)
        execCmd.Stdout = os.Stdout
        execCmd.Stderr = os.Stderr
        execCmd.Run()
    },
}
```

## 3. Designing Interactive UIs

The Go version uses the **Bubble Tea** framework. Instead of calling `gum`, use the pre-built components in the `internal/ui` package.

### Confirmation with Timeout
```go
confirmed, _ := ui.Confirm("Proceed with deploy?", 5*time.Second, true)
```

### Filtering Selection
```go
choice, _ := ui.Choose("Select environment:", []string{"dev", "staging", "prod"})
```

## 4. Rebuilding
After making changes to the Go code, you must rebuild the binary:

```bash
go build -o minhthetus-cli main.go
```
