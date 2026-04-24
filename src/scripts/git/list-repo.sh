#!/bin/bash
# Description: Interactively manage and list tracked repositories.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Repository List Manager"
HELP_USAGE="minhthetus-cli git list-repo"
HELP_DESCRIPTION="Interactively view, add, or remove repositories from the tracking list used for bulk scans."

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

set -euo pipefail

LIST_REPO_FILE="$HOME/.minhthetus-cli/list-repo.json"

# Function to ensure the file exists
ensure_config() {
    if [[ ! -f "$LIST_REPO_FILE" ]]; then
        mkdir -p "$(dirname "$LIST_REPO_FILE")"
        echo "[]" > "$LIST_REPO_FILE"
    fi
}

# Function to add a repository
add_new_repo() {
    printf "\n%b\n" "${TAG} ${BOLD}Registering New Repository${NC}"
    local path=$(gum input --placeholder "Enter absolute path to repository (e.g. /Users/me/project)...")
    if [[ -z "$path" ]]; then return; fi
    
    # Check if directory exists
    if [[ ! -d "$path" ]]; then
        printf "%b\n" "${ERROR} ${RED}Directory does not exist: $path${NC}"
        gum input --placeholder "Press Enter to continue..." --value "" > /dev/null || true
        return
    fi
    
    # Get absolute path
    local abs_path=$(cd "$path" && pwd)
    
    # Call the CLI to track
    minhthetus-cli repo-track "$abs_path"
    printf "%b\n" "${CHECK} ${GREEN}Repository added.${NC}"
    gum input --placeholder "Press Enter to continue..." --value "" > /dev/null || true
}

# Main Interactive Loop
while true; do
    ensure_config
    clear
    
    repos_count=$(jq '. | length' "$LIST_REPO_FILE")
    
    if [[ "$repos_count" -eq 0 ]]; then
        printf "%b\n" "${YELLOW}${INFO} No repositories are currently being tracked.${NC}"
        action=$(gum choose "Add New" "Quit")
    else
        # Prepare list for gum filter
        # Format: Name | Path
        options=$(jq -r '.[] | "\(.name) | \(.path)"' "$LIST_REPO_FILE")
        add_opt="➕ Add New"
        quit_opt="🚪 Quit"
        
        choice=$(printf "%s\n%s\n%s" "$options" "$add_opt" "$quit_opt" | gum filter --placeholder "Select repository to manage...")
        
        if [[ -z "$choice" || "$choice" == "$quit_opt" ]]; then
            exit 0
        elif [[ "$choice" == "$add_opt" ]]; then
            add_new_repo
            continue
        fi
        
        # Extract path from choice (everything after the | )
        repo_path=$(echo "$choice" | awk -F ' | ' '{print $NF}')
        
        # Show repository details
        clear
        printf "%b\n" "${TAG} ${BOLD}Repository Details:${NC}"
        printf "\n"
        jq -r ".[] | select(.path == \"$repo_path\") | \"  ${BOLD}Name:${NC}        \(.name)\n  ${BOLD}Description:${NC} \(.description)\n  ${BOLD}Path:${NC}        \(.path)\"" "$LIST_REPO_FILE"
        printf "\n"
        
        sub_action=$(gum choose "Delete" "Back")
        if [[ "$sub_action" == "Delete" ]]; then
            if gum confirm "Are you sure you want to untrack this repository?"; then
                minhthetus-cli repo-untrack "$repo_path"
                printf "%b\n" "${CHECK} ${GREEN}Repository removed from tracking list.${NC}"
                gum input --placeholder "Press Enter to continue..." --value "" > /dev/null || true
            fi
        fi
        continue
    fi
    
    case "$action" in
        "Add New") add_new_repo ;;
        "Quit"|"") exit 0 ;;
    esac
done
