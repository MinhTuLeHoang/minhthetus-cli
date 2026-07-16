//go:build dev

package cmd

import (
	"fmt"

	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/debug"
)

func init() {
	fmt.Println("\n[DEV MODE]\n")
	registerDebug = func() {
		rootCmd.AddCommand(debug.Cmd)
	}
}
