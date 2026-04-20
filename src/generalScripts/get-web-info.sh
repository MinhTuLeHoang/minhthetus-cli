#!/bin/bash
# Description: Detects Node version (via .nvmrc or nvm) and package manager (pnpm/npm/yarn).

# Source constants and utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/constants.sh"
source "$SCRIPT_DIR/emit-shell-command.sh"

# Path to gum
GUM="$SCRIPT_DIR/../../vendor/bin/gum"

# Function to load nvm
load_nvm() {
    if [ -s "$HOME/.nvm/nvm.sh" ]; then
        . "$HOME/.nvm/nvm.sh"
    elif [ -n "$(command -v brew)" ] && [ -f "$(brew --prefix nvm)/nvm.sh" ]; then
        . "$(brew --prefix nvm)/nvm.sh"
    fi
}

# STEP 1: Node version
NODE_VERSION=""
if [ -f ".nvmrc" ]; then
    NODE_VERSION=$(cat .nvmrc | tr -d '[:space:]')
    echo -e "${GREEN}${CHECK} Found .nvmrc: $NODE_VERSION${NC}" >&2
    emit_shell_command "nvm use"
else
    echo -e "${YELLOW}${INFO} No .nvmrc found. Detecting Node versions...${NC}" >&2
    load_nvm
    if command -v nvm >/dev/null 2>&1; then
        # Parse nvm ls output
        RAW_VERSIONS=$(nvm ls --no-colors)
        # Extract things like v20.10.0
        VERSIONS=$(echo "$RAW_VERSIONS" | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | sort -Vr | uniq)
        
        if [ -z "$VERSIONS" ]; then
             # Try a different approach for versions if first one fails
             VERSIONS=$(echo "$RAW_VERSIONS" | grep -E "^[[:space:]]*v[0-9]" | sed 's/[[:space:]]*//' | cut -f1 -d' ' | sort -Vr)
        fi

        if [ -n "$VERSIONS" ]; then
            SELECTED_VERSION=$(echo "$VERSIONS" | "$GUM" choose --header "Select Node version:")
            if [ -n "$SELECTED_VERSION" ]; then
                NODE_VERSION=$SELECTED_VERSION
                # Create .nvmrc with major version
                MAJOR_VERSION=$(echo "$NODE_VERSION" | grep -o 'v[0-9]\+' | head -n 1)
                echo "$MAJOR_VERSION" > .nvmrc
                echo -e "${GREEN}${CHECK} Created .nvmrc with $MAJOR_VERSION${NC}" >&2
                emit_shell_command "nvm use $NODE_VERSION"
            fi
        else
            echo -e "${RED}${ERROR} No Node versions found via nvm.${NC}" >&2
            NODE_VERSION=$(node -v 2>/dev/null || echo "unknown")
        fi
    else
        echo -e "${RED}${ERROR} nvm not found. Using current node version.${NC}" >&2
        NODE_VERSION=$(node -v 2>/dev/null || echo "unknown")
    fi
fi

# STEP 2: detect npm / pnpm / yarn
PACKAGE_MANAGER=""
LOCK_FILES=()
[ -f "pnpm-lock.yaml" ] && LOCK_FILES+=("pnpm")
[ -f "package-lock.json" ] && LOCK_FILES+=("npm")
[ -f "yarn.lock" ] && LOCK_FILES+=("yarn")

NUM_LOCKS=${#LOCK_FILES[@]}

if [ $NUM_LOCKS -eq 0 ]; then
    echo -e "${YELLOW}${INFO} No lock files found. Please choose a package manager.${NC}" >&2
    PACKAGE_MANAGER=$("$GUM" choose "pnpm" "npm" "yarn" --header "Select Package Manager:")
elif [ $NUM_LOCKS -eq 1 ]; then
    PACKAGE_MANAGER=${LOCK_FILES[0]}
    echo -e "${GREEN}${CHECK} Detected package manager: $PACKAGE_MANAGER${NC}" >&2
else
    echo -e "${YELLOW}${INFO} Multiple lock files detected.${NC}" >&2
    PACKAGE_MANAGER=$(printf "%s\n" "${LOCK_FILES[@]}" | "$GUM" choose --header "Select Package Manager to use:")
    
    # Ask to delete others
    for lock in "${LOCK_FILES[@]}"; do
        if [ "$lock" != "$PACKAGE_MANAGER" ]; then
            FILE_TO_DELETE=""
            [ "$lock" == "pnpm" ] && FILE_TO_DELETE="pnpm-lock.yaml"
            [ "$lock" == "npm" ] && FILE_TO_DELETE="package-lock.json"
            [ "$lock" == "yarn" ] && FILE_TO_DELETE="yarn.lock"
            
            if "$GUM" confirm "Delete redundant $FILE_TO_DELETE?"; then
                rm "$FILE_TO_DELETE"
                echo -e "${GREEN}${CHECK} Deleted $FILE_TO_DELETE${NC}" >&2
            fi
        fi
    done
fi

# Final Output JSON - ONLY this should go to stdout
echo "{\"nodeVersion\": \"$NODE_VERSION\", \"packageManager\": \"$PACKAGE_MANAGER\"}"

