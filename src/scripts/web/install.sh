#!/bin/bash
# Description: Smart web project dependency installer

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Web Project Installer"

HELP_USAGE="minhthetus-cli web install [options]"
HELP_DESCRIPTION="Installs project dependencies with automatic Node.js version switching and package manager detection."
HELP_OPTIONS="-f, --force | Force install: removes node_modules and existing lock files before installing.
--ci | CI mode: installs dependencies using the frozen lockfile."

HELP_EXAMPLE="minhthetus-cli web install --force\nminhthetus-cli web install --ci"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"
source "$GENERAL_SCRIPTS_DIR/get-web-info.sh"

# Background: Track this repository silently
minhthetus-cli repo-track "$(pwd)" --silent &> /dev/null &


# Parse arguments
FORCE=false
CI_MODE=false
for arg in "$@"; do
    case $arg in
        -f|--force)
            FORCE=true
            shift
            ;;
        --ci)
            CI_MODE=true
            shift
            ;;
    esac
done

# Step 1: Detect environment
printf "%b\n" "${BLUE}${INFO} Detecting environment...${NC}"
printf "\n"
get_web_info

if [ -z "$G_PACKAGE_MANAGER" ]; then
    printf "\n"
    printf "%b\n" "${RED}${ERROR} Failed to detect package manager.${NC}"
    exit 1
fi

# Step 2: Handle Force Install
if [ "$FORCE" = true ]; then
    printf "\n"
    printf "%b\n" "${YELLOW}${WARNING} Force mode enabled. Cleaning up...${NC}"
    
    if [ -d "node_modules" ]; then
        printf "%b\n" "${_INDENT}${INFO} Removing node_modules...${NC}"
        rm -rf node_modules
    fi

    # Remove lock files based on detected package manager (or all if we want to be thorough)
    case $G_PACKAGE_MANAGER in
        pnpm)
            [ -f "pnpm-lock.yaml" ] && rm "pnpm-lock.yaml" && printf "%b\n" "${_INDENT}${INFO} Removed pnpm-lock.yaml${NC}"
            ;;
        npm)
            [ -f "package-lock.json" ] && rm "package-lock.json" && printf "%b\n" "${_INDENT}${INFO} Removed package-lock.json${NC}"
            ;;
        yarn)
            [ -f "yarn.lock" ] && rm "yarn.lock" && printf "%b\n" "${_INDENT}${INFO} Removed yarn.lock${NC}"
            ;;
    esac
fi

# Step 3: Execute Install
printf "\n"
INSTALL_MSG="${BLUE}${ROCKET} Installing dependencies using ${BOLD}${G_PACKAGE_MANAGER}${NC}"
START_MS=$(get_time_ms)

case $G_PACKAGE_MANAGER in
    pnpm)
        if [ "$CI_MODE" = true ]; then pnpm install --frozen-lockfile; else pnpm i; fi
        ;;
    npm)    
        if [ "$CI_MODE" = true ]; then npm ci; else npm i; fi
        ;;
    yarn)
        if [ "$CI_MODE" = true ]; then yarn install --frozen-lockfile; else yarn install; fi
        ;;
esac

INSTALL_EXIT_CODE=$?
END_MS=$(get_time_ms)
ELAPSED=$((END_MS - START_MS))
SEC=$((ELAPSED / 1000))
MS=$(((ELAPSED % 1000) / 100))
G_DURATION="${SEC}.${MS}s"

# Final status
printf "\n"
if [ $INSTALL_EXIT_CODE -eq 0 ]; then
    printf "%b\n" "${INSTALL_MSG} ${GREEN}${CHECK}${NC} (${G_DURATION})"
else
    printf "%b\n" "${RED}${ERROR} Installation failed with exit code ${INSTALL_EXIT_CODE}${NC} (${G_DURATION})"
    exit 1
fi
