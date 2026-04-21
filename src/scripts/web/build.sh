#!/bin/bash
# Description: Builds the web project with automatic environment detection.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Web Project Builder"
HELP_USAGE="minhthetus-cli web build [options] [-- [args]]"
HELP_DESCRIPTION="Detects the environment and runs the build script using the appropriate package manager."
HELP_OPTIONS="[args] | Pass additional arguments to the build command."

HELP_EXAMPLE="minhthetus-cli web build"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"
source "$GENERAL_SCRIPTS_DIR/get-web-info.sh"

# Step 1: Detect environment
printf "%b\n" "${BLUE}${INFO} Detecting environment...${NC}"
get_web_info

if [ -z "$G_PACKAGE_MANAGER" ]; then
    printf "%b\n" "${RED}${ERROR} Failed to detect package manager.${NC}"
    exit 1
fi

# Step 2: Prepare and execute build command
printf "%b\n" "${BLUE}${ROCKET} Building project using ${BOLD}${G_PACKAGE_MANAGER}${NC}..."


case $G_PACKAGE_MANAGER in
    pnpm)
        pnpm run build "$@"
        ;;
    npm)
        npm run build "$@"
        ;;
    yarn)
        yarn run build "$@"
        ;;
    *)
        printf "%b\n" "${RED}${ERROR} Unsupported package manager: $G_PACKAGE_MANAGER${NC}"
        exit 1
        ;;
esac