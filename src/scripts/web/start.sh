#!/bin/bash
# Description: Starts the web project development server with automatic environment detection.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Web Project Starter"
HELP_USAGE="minhthetus-cli web start [options] [-- [args]]"
HELP_DESCRIPTION="Detects the environment and starts the development server using the appropriate package manager."
HELP_OPTIONS="[args]| Pass additional arguments to the start command."

HELP_EXAMPLE="minhthetus-cli web start --port 3000"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"
source "$GENERAL_SCRIPTS_DIR/get-web-info.sh"

# Step 1: Detect environment
printf "%b\n" "${BLUE}${INFO} Detecting environment...${NC}"
get_web_info

if [ -z "$G_PACKAGE_MANAGER" ]; then
    printf "\n"
    printf "%b\n" "${RED}${ERROR} Failed to detect package manager.${NC}"
    exit 1
fi

# Step 2: Prepare and execute start command
printf "\n"
printf "%b\n" "${BLUE}${ROCKET} Starting project using ${BOLD}${G_PACKAGE_MANAGER}${NC}..."


case $G_PACKAGE_MANAGER in
    pnpm)
        pnpm start "$@"
        ;;
    npm)
        npm start "$@"
        ;;
    yarn)
        yarn start "$@"
        ;;
    *)
        printf "%b\n" "${RED}${ERROR} Unsupported package manager: $G_PACKAGE_MANAGER${NC}"
        exit 1
        ;;
esac