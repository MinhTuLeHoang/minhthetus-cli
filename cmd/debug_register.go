//go:build dev

package cmd

import (
	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/debug"
)

func init() {
	registerDebug = func() {
		rootCmd.AddCommand(debug.Cmd)
	}
}
