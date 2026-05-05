#!/bin/bash
# Description: Publish project code to the remote server

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Publish Code"
HELP_USAGE="minhthetus-cli vng publish-code"
HELP_DESCRIPTION="Publishes the project codebase to the remote VNG server."

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

echo "Publishing code..."
