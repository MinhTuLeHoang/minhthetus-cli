#!/bin/bash
# Description: Smart web project dependency installer

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Web Project Installer"

HELP_USAGE="minhthetus-cli web install [options]"
HELP_DESCRIPTION="Installs project dependencies with automatic Node.js version switching and package manager detection."
HELP_OPTIONS="-f, --force      | Force install: removes node_modules and existing lock files before installing."

HELP_EXAMPLE="minhthetus-cli web install --force"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"
source "$GENERAL_SCRIPTS_DIR/get-web-info.sh"

# Parse arguments
FORCE=false
for arg in "$@"; do
    case $arg in
        -f|--force)
            FORCE=true
            shift
            ;;
    esac
done

# Step 1: Detect environment
printf "%b\n" "${BLUE}${INFO} Detecting environment...${NC}"
get_web_info

if [ -z "$G_PACKAGE_MANAGER" ]; then
    printf "%b\n" "${RED}${ERROR} Failed to detect package manager.${NC}"
    exit 1
fi

# Step 2: Handle Force Install
if [ "$FORCE" = true ]; then

    printf "%b\n" "${YELLOW}${WARNING} Force mode enabled. Cleaning up...${NC}"

    
    if [ -d "node_modules" ]; then
        printf "%b\n" "${_INDENT:-  }${INFO} Removing node_modules...${NC}"
        rm -rf node_modules
    fi

    # Remove lock files based on detected package manager (or all if we want to be thorough)
    case $G_PACKAGE_MANAGER in
        pnpm)
            [ -f "pnpm-lock.yaml" ] && rm "pnpm-lock.yaml" && printf "%b\n" "${_INDENT:-  }${INFO} Removed pnpm-lock.yaml${NC}"
            ;;
        npm)
            [ -f "package-lock.json" ] && rm "package-lock.json" && printf "%b\n" "${_INDENT:-  }${INFO} Removed package-lock.json${NC}"
            ;;
        yarn)
            [ -f "yarn.lock" ] && rm "yarn.lock" && printf "%b\n" "${_INDENT:-  }${INFO} Removed yarn.lock${NC}"
            ;;
    esac
fi

# Step 3: Execute Install
printf "%b\n" "${BLUE}${ROCKET} Installing dependencies using ${BOLD}${G_PACKAGE_MANAGER}${NC}..."


case $G_PACKAGE_MANAGER in
    pnpm)
        if [ "$FORCE" = true ] || [ ! -f "pnpm-lock.yaml" ]; then
            pnpm install
        else
            pnpm install --frozen-lockfile
        fi
        ;;
    npm)
        if [ "$FORCE" = true ] || [ ! -f "package-lock.json" ]; then
            npm install
        else
            npm ci
        fi
        ;;
    yarn)
        if [ "$FORCE" = true ] || [ ! -f "yarn.lock" ]; then
            yarn install
        else
            yarn install --frozen-lockfile
        fi
        ;;
    *)
        printf "%b\n" "${RED}${ERROR} Unsupported package manager: $G_PACKAGE_MANAGER${NC}"
        exit 1
        ;;
esac

if [ $? -eq 0 ]; then
    printf "%b\n" "${GREEN}${CHECK} Dependencies installed successfully!${NC}"
else
    printf "%b\n" "${RED}${ERROR} Installation failed.${NC}"
    exit 1
fi