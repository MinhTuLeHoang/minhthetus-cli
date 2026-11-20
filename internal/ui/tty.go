package ui

import (
	"errors"
	"os"

	"github.com/mattn/go-isatty"
)

// ErrNonInteractiveTTY is returned when an interactive prompt is invoked without a terminal (TTY).
var ErrNonInteractiveTTY = errors.New("interactive prompt requires a terminal (TTY)")

func isTTYFd(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func isInteractiveTTY() bool {
	return isTTYFd(os.Stdin.Fd()) && isTTYFd(os.Stdout.Fd())
}
