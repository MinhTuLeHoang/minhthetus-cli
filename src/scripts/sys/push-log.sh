#!/bin/bash
# Description: Push system logs to remote

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Push Logs"
HELP_USAGE="minhthetus-cli sys push-log"
HELP_DESCRIPTION="Pushes system-level log files to a remote storage."

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

echo "Push logs..." 
