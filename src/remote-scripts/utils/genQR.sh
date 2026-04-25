#!/bin/bash

# Description: Generate a QR code from text or URL in your terminal.

# Help metadata
HELP_TITLE="QR Generator"
HELP_USAGE="minhthetus-cli utils genQR <content> [options]"
HELP_DESCRIPTION="Generates a high-quality QR code directly in your terminal.
This command uses the qrenco.de service to keep the local CLI package lightweight."
HELP_OPTIONS="-m, --mode <type> | Generation mode: 'api' (default) or 'self-implement'"
HELP_EXAMPLE="minhthetus-cli utils genQR 'hello'\nminhthetus-cli utils genQR 'hello' --mode self-implement"

# Source help system
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../../generalScripts/print-help.sh" "$@"

set -euo pipefail

# Parse arguments
CONTENT=""
MODE="api"

while [[ $# -gt 0 ]]; do
    case $1 in
        --mode|-m)
            MODE="$2"
            shift 2
            ;;
        *)
            if [ -z "$CONTENT" ]; then
                CONTENT="$1"
            fi
            shift
            ;;
    esac
done

if [ -z "$CONTENT" ]; then
    printf "\n%b %b\n\n" "${ERROR}" "${RED}Error: Content is required.${NC}"
    printf "Usage: minhthetus-cli utils genQR <content>\n"
    exit 1
fi

if [ "$MODE" = "api" ]; then
    printf "\n%b %b\n\n" "${INFO}" "${CYAN}Generating QR Code (via qrenco.de API)...${NC}"
    echo "$CONTENT" | curl -s -F "-=<-" "https://qrenco.de/"
else
    printf "\n%b %b\n\n" "${INFO}" "${CYAN}Generating QR Code (Local Engine)...${NC}"
    # Use self-implemented Node.js script
    NODE_SCRIPT="$SCRIPT_DIR/qr-logic.js"
    node "$NODE_SCRIPT" "$CONTENT"
fi

printf "\n"
printf "%b %b\n" "${CHECK}" "${GREEN}QR Code generated successfully!${NC}"

printf "\n"
