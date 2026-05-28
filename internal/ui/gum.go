package ui

import (
	"os"
	"os/exec"
	"strings"
)

// GumChoose opens a gum choose menu
func GumChoose(options ...string) string {
	cmd := exec.Command("gum", append([]string{"choose"}, options...)...)
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

// GumConfirm opens a gum confirm dialog
func GumConfirm(prompt string) bool {
	cmd := exec.Command("gum", "confirm", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}

// GumConfirmTimeout opens a gum confirm dialog with a timeout
func GumConfirmTimeout(prompt, timeout string) bool {
	cmd := exec.Command("gum", "confirm", prompt, "--timeout", timeout, "--default", "Yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err == nil
}

// GumInput opens a gum input field
func GumInput(placeholder, value string) string {
	args := []string{"input", "--placeholder", placeholder}
	if value != "" {
		args = append(args, "--value", value)
	}
	cmd := exec.Command("gum", args...)
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

// GumFilter opens a gum filter menu
func GumFilter(options []string, placeholder string) string {
	input := strings.Join(options, "\n")
	cmd := exec.Command("gum", "filter", "--placeholder", placeholder)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}
