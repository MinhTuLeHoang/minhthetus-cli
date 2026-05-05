#!/bin/bash
# Description: Clean up system logs

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Clean Logs"
HELP_USAGE="minhthetus-cli sys clean-log"
HELP_DESCRIPTION="Cleans up system-level log files."

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

echo "Cleaning logs..." 
