#!/bin/bash
# Description: Detects Node version (via .nvmrc or nvm) and package manager (pnpm/npm/yarn).

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/emit-shell-command.sh"

# Path to gum
GUM="gum"
if ! command -v gum >/dev/null 2>&1; then
    GUM="$SCRIPT_DIR/../../vendor/bin/gum"
fi

# Function to load nvm into the current shell session
load_nvm() {
    if [ -s "$HOME/.nvm/nvm.sh" ]; then
        . "$HOME/.nvm/nvm.sh"
    elif [ -n "$(command -v brew)" ] && [ -f "$(brew --prefix nvm)/nvm.sh" ]; then
        . "$(brew --prefix nvm)/nvm.sh"
    fi
}

# Function to detect environment info
get_web_info() {
    load_nvm

    # STEP 1: Node version
    G_NODE_VERSION=""
    if [ -f ".nvmrc" ]; then
        G_NODE_VERSION=$(cat .nvmrc | tr -d '[:space:]')
        printf "%b\n" "${GREEN}${CHECK} Found .nvmrc: $G_NODE_VERSION${NC}" >&2
        emit_shell_command "nvm use --silent"
        
        # Also switch in the current shell session so subsequent commands use it
        if command -v nvm >/dev/null 2>&1; then
            nvm use "$G_NODE_VERSION"
        fi
    else
        printf "%b\n" "${YELLOW}${INFO} No .nvmrc found.${NC}" >&2
        if command -v nvm >/dev/null 2>&1; then
            local RAW_VERSIONS=""
            local START_TIME=$(date +%s)

            # Fast Path: Direct directory listing (instant)
            local NVM_DIR_LOCAL="${NVM_DIR:-$HOME/.nvm}"
            if [ -d "$NVM_DIR_LOCAL/versions/node" ]; then
                RAW_VERSIONS=$(ls -1 "$NVM_DIR_LOCAL/versions/node" 2>/dev/null)
            fi

            # Fallback: Slow nvm list (only if directory structure is different)
            if [ -z "$RAW_VERSIONS" ]; then
                RAW_VERSIONS=$(nvm list --no-colors)
            fi

            local DURATION=$(( $(date +%s) - START_TIME ))
            if [ $DURATION -gt 0 ]; then
                printf "%b\n" "${GREEN}${CHECK} Fetched in ${DURATION}s${NC}" >&2
            fi
            # Extract things like v20.10.0
            # Extract only the local version numbers (lines starting with space/arrow and 'v')
            local VERSIONS=$(echo "$RAW_VERSIONS" | grep -E '^[[:space:]]*(->)?[[:space:]]*v[0-9]' | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]' | sort -Vr | uniq)
            
            if [ -z "$VERSIONS" ]; then
                 VERSIONS=$(echo "$RAW_VERSIONS" | grep -E '^[[:space:]]*(->)?[[:space:]]*v[0-9]' | sed -E 's/^[[:space:]]*(->)?[[:space:]]*//' | cut -f1 -d' ' | sort -Vr)
            fi

            if [ -n "$VERSIONS" ]; then
                local SELECTED_VERSION=$(echo "$VERSIONS" | "$GUM" choose --header "Select Node version:")
                if [ -n "$SELECTED_VERSION" ]; then
                    G_NODE_VERSION=$SELECTED_VERSION
                    MAJOR_VERSION=$(echo "$G_NODE_VERSION" | grep -o 'v[0-9]\+' | head -n 1)
                    echo "$MAJOR_VERSION" > .nvmrc
                    printf "%b\n" "${GREEN}${CHECK} Created .nvmrc with $MAJOR_VERSION${NC}" >&2
                    emit_shell_command "nvm use $G_NODE_VERSION --silent"
                    nvm use "$G_NODE_VERSION"
                fi
            else
                printf "%b\n" "${RED}${ERROR} No Node versions found via nvm.${NC}" >&2
                G_NODE_VERSION=$(node -v 2>/dev/null || echo "unknown")
            fi
        else
            printf "%b\n" "${RED}${ERROR} nvm not found. Using current node version.${NC}" >&2
            G_NODE_VERSION=$(node -v 2>/dev/null || echo "unknown")
        fi
    fi

    # STEP 2: detect npm / pnpm / yarn
    G_PACKAGE_MANAGER=""
    local LOCK_FILES=()
    [ -f "pnpm-lock.yaml" ] && LOCK_FILES+=("pnpm")
    [ -f "package-lock.json" ] && LOCK_FILES+=("npm")
    [ -f "yarn.lock" ] && LOCK_FILES+=("yarn")

    local NUM_LOCKS=${#LOCK_FILES[@]}

    if [ $NUM_LOCKS -eq 0 ]; then
        printf "\n"
        printf "%b\n" "${YELLOW}${WARNING} No lock files found. Please choose a package manager.${NC}" >&2
        G_PACKAGE_MANAGER=$("$GUM" choose "pnpm" "npm" "yarn" --header "Select Package Manager:")
    elif [ $NUM_LOCKS -eq 1 ]; then
        G_PACKAGE_MANAGER=${LOCK_FILES[0]}
        printf "%b\n" "${GREEN}${CHECK} Detected package manager: $G_PACKAGE_MANAGER${NC}" >&2
    else
        printf "%b\n" "${YELLOW}${INFO} Multiple lock files detected.${NC}" >&2
        G_PACKAGE_MANAGER=$(printf "%s\n" "${LOCK_FILES[@]}" | "$GUM" choose --header "Select Package Manager to use:")
        
        for lock in "${LOCK_FILES[@]}"; do
            if [ "$lock" != "$G_PACKAGE_MANAGER" ]; then
                local FILE_TO_DELETE=""
                [ "$lock" == "pnpm" ] && FILE_TO_DELETE="pnpm-lock.yaml"
                [ "$lock" == "npm" ] && FILE_TO_DELETE="package-lock.json"
                [ "$lock" == "yarn" ] && FILE_TO_DELETE="yarn.lock"
                
                if "$GUM" confirm "Delete redundant $FILE_TO_DELETE?"; then
                    rm "$FILE_TO_DELETE"
                    printf "%b\n" "${GREEN}${CHECK} Deleted $FILE_TO_DELETE${NC}" >&2
                fi
            fi
        done
    fi

    # Print JSON only if run as a script
    if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
        echo "{\"nodeVersion\": \"$G_NODE_VERSION\", \"packageManager\": \"$G_PACKAGE_MANAGER\"}"
    fi
}


# If script is run directly, execute the function
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    get_web_info
fi


