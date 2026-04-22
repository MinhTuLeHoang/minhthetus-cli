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
    printf "\n"
    printf "%b\n" "${RED}${ERROR} Failed to detect package manager.${NC}"
    exit 1
fi

# Step 2: Prepare and execute build command
printf "\n"
BUILD_MSG="${BLUE}${ROCKET} Building project using ${BOLD}${G_PACKAGE_MANAGER}${NC}"
LOG_FILE=$(mktemp)

# Execute build in background
(
    case $G_PACKAGE_MANAGER in
        pnpm) pnpm run build "$@" ;;
        npm)  npm run build "$@" ;;
        yarn) yarn run build "$@" ;;
    esac
) > "$LOG_FILE" 2>&1 &

BUILD_PID=$!

# Show spinner while building
show_spinner "$BUILD_PID" "$BUILD_MSG" "$LOG_FILE"

# Wait for process and get exit code
wait "$BUILD_PID"
BUILD_EXIT_CODE=$?

# Clear spinner line and show final status
if [ $BUILD_EXIT_CODE -eq 0 ]; then
    # We need to recalculate or just print success. show_spinner left the cursor at the end of last line.
    printf "\r${_INDENT:-  }${BUILD_MSG} ${GREEN}${CHECK}${NC} (${G_DURATION})\n"
    rm "$LOG_FILE"
else
    printf "\r${_INDENT:-  }${BUILD_MSG} ${RED}${ERROR}${NC} (${G_DURATION})\n"
    printf "\n"
    printf "%b\n" "${RED}${ERROR} Build failed with exit code ${BUILD_EXIT_CODE}${NC}"
    rm "$LOG_FILE"
    exit $BUILD_EXIT_CODE
fi