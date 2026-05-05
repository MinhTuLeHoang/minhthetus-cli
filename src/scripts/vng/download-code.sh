#!/bin/bash
# Description: Download project code from the remote server

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Download Code"
HELP_USAGE="minhthetus-cli vng download-code"
HELP_DESCRIPTION="Downloads the project codebase from the remote VNG server."

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

echo "Download code..."
