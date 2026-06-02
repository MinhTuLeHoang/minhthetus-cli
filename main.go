/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

🤖 IMPORTANT AI AGENT DEVELOPER SAFEGUARD:
DO NOT run raw "go build", "go run", "go test", or "go clean" directly!
Direct "go" commands are blocked by the sandbox and prompt the user for permission.
Instead, ALWAYS compile using the pre-approved "make" commands:
  - Build binary:  make build
  - Build debug:   make build-dev
  - Run binary:    ./minhthetus-cli <args>
*/
package main

import "github.com/MinhTuLeHoang/minhthetus-cli/cmd"

func main() {
	cmd.Execute()
}
