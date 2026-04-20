#!/bin/bash
# Description: Simple greeting script that accepts a --name parameter.

# Source constants and utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../generalScripts/constants.sh"

# Define help metadata
HELP_TITLE="Hello"
HELP_USAGE="./hello.sh [options]"
HELP_DESCRIPTION="A simple greeting script that welcomes you by name."
HELP_OPTIONS="--name <name> | The name to greet (default: User)
-h, --help     | Show this help message"
HELP_EXAMPLE="./hello.sh --name Antigravity"

# Check for help flags
source "$SCRIPT_DIR/../generalScripts/print-help.sh" "$@"

# Simple parameter parsing for --name
NAME="User" # Default name

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --name)
      NAME="$2"
      shift
      ;;
  esac
  shift
done

printf "%b\n" "${GREEN}${CHECK} Hello $NAME!${NC}"
