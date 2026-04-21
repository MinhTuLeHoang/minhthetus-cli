#!/bin/bash
# Description: Advanced Git account manager. Supports switching identities and managing saved accounts (create/delete).

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Git Account Manager"
HELP_USAGE="minhthetus-cli git account [options]"
HELP_DESCRIPTION="Managed Git identities. Quickly switch between accounts or manage your saved list."
HELP_OPTIONS="-m, --manage     | Enter management mode (list, create, delete accounts)"
HELP_EXAMPLE="minhthetus-cli git account --manage"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

set -euo pipefail

# --- Configuration ---
CONFIG_DIR="$HOME/.minhthetus-cli"
CONFIG_FILE="$CONFIG_DIR/git-accounts.json"

mkdir -p "$CONFIG_DIR"
if [[ ! -f "$CONFIG_FILE" || ! -s "$CONFIG_FILE" ]]; then
    echo "[]" > "$CONFIG_FILE"
fi

# --- Utilities ---
detect_identity() {
    local levels=("local" "global" "system")
    local found_email=""
    local found_name=""
    local email_level=""
    local name_level=""

    for level in "${levels[@]}"; do
        if [[ -z "$found_email" ]]; then
            val=$(git config --$level user.email 2>/dev/null || true)
            if [[ -n "$val" ]]; then
                found_email="$val"
                email_level="$level"
            fi
        fi
        if [[ -z "$found_name" ]]; then
            val=$(git config --$level user.name 2>/dev/null || true)
            if [[ -n "$val" ]]; then
                found_name="$val"
                name_level="$level"
            fi
        fi
    done

    printf "%b\n" "${TAG} ${BOLD}Current Identity Detection:${NC}"
    if [[ -n "$found_email" ]]; then
        printf "  Email: ${CYAN}%-30s${NC} (from ${YELLOW}%s${NC})\n" "$found_email" "$email_level"
    else
        printf "  Email: ${RED}Not set${NC}\n"
    fi
    
    if [[ -n "$found_name" ]]; then
        printf "  Name:  ${CYAN}%-30s${NC} (from ${YELLOW}%s${NC})\n" "$found_name" "$name_level"
    else
        printf "  Name:  ${RED}Not set${NC}\n"
    fi
}

list_accounts() {
    printf "\n%b\n" "${TAG} ${BOLD}Saved Accounts:${NC}"
    local accounts_count=$(jq '. | length' "$CONFIG_FILE")
    if [[ "$accounts_count" -eq 0 ]]; then
        printf "  ${YELLOW}No accounts saved.${NC}\n"
    else
        jq -r '["TITLE", "NAME", "EMAIL"], ["-----", "----", "-----"], (.[] | [.title, .name, .email]) | @tsv' "$CONFIG_FILE" | column -t -s $'\t' | sed 's/^/  /'
    fi
    printf "\n"
}

save_account() {
    local title="$1"
    local name="$2"
    local email="$3"
    local tmp_file=$(mktemp)
    jq ". += [{\"title\": \"$title\", \"name\": \"$name\", \"email\": \"$email\"}]" "$CONFIG_FILE" > "$tmp_file"
    mv "$tmp_file" "$CONFIG_FILE"
}

delete_account() {
    local options=$(jq -r '.[] | "\(.title) ( \(.name) <\(.email)> )"' "$CONFIG_FILE")
    if [[ -z "$options" ]]; then
        printf "%b\n" "${WARNING} ${YELLOW}No accounts to delete.${NC}"
        return
    fi
    
    local choice=$(echo "$options" | gum filter --placeholder "Select account to DELETE...")
    if [[ -z "$choice" ]]; then return; fi
    
    local title=$(echo "$choice" | sed -E 's/ \(.*//')
    if gum confirm "Are you sure you want to delete '$title'?"; then
        local tmp_file=$(mktemp)
        jq "del(.[] | select(.title == \"$title\"))" "$CONFIG_FILE" > "$tmp_file"
        mv "$tmp_file" "$CONFIG_FILE"
        printf "%b\n" "${CHECK} ${GREEN}Account deleted.${NC}"
    fi
}

create_new_account() {
    printf "\n%b\n" "${TAG} ${BOLD}Registering New Account${NC}"
    local title=$(gum input --placeholder "Title (e.g. Work)")
    if [[ -z "$title" ]]; then return; fi
    
    local name=$(gum input --placeholder "User Name")
    local email=$(gum input --placeholder "User Email")
    
    if [[ -n "$name" && -n "$email" ]]; then
        save_account "$title" "$name" "$email"
        printf "%b\n" "${CHECK} ${GREEN}Account saved.${NC}"
    else
        printf "%b\n" "${ERROR} ${RED}Invalid input. Required all fields.${NC}"
    fi
}

# --- Argument Parsing ---
MANAGE_MODE=false
while [[ $# -gt 0 ]]; do
    case "$1" in
        -m|--manage) MANAGE_MODE=true; shift ;;
        *) shift ;;
    esac
done

# --- MAIN LOGIC ---
if [[ "$MANAGE_MODE" == "true" ]]; then
    while true; do
        clear
        detect_identity
        list_accounts
        
        action=$(gum choose "Create New" "Delete" "Quit")
        case "$action" in
            "Create New") create_new_account ;;
            "Delete") delete_account ;;
            "Quit"|"") exit 0 ;;
        esac
        
        printf "\n"
        gum input --placeholder "Press Enter to continue..." --value "" > /dev/null || true
    done
else
    detect_identity
    printf "\n"
    
    if ! gum confirm "Switch account for this repository?"; then
        exit 0
    fi
    
    accounts_count=$(jq '. | length' "$CONFIG_FILE")
    if [[ "$accounts_count" -eq 0 ]]; then
        create_new_account
    else
        options=$(jq -r '.[] | "\(.title) ( \(.name) <\(.email)> )"' "$CONFIG_FILE")
        add_opt="➕ Add New Account"
        quit_opt="🚪 Quit"
        
        choice=$(printf "%s\n%s\n%s" "$options" "$add_opt" "$quit_opt" | gum filter --placeholder "Select identity to apply...")
        
        if [[ -z "$choice" || "$choice" == "$quit_opt" ]]; then
            exit 0
        elif [[ "$choice" == "$add_opt" ]]; then
            create_new_account
        else
            title=$(echo "$choice" | sed -E 's/ \(.*//')
            selected_json=$(jq -c ".[] | select(.title == \"$title\")" "$CONFIG_FILE" | head -1)
            SELECTED_NAME=$(echo "$selected_json" | jq -r '.name')
            SELECTED_EMAIL=$(echo "$selected_json" | jq -r '.email')
            
            git config user.email "$SELECTED_EMAIL"
            git config user.name "$SELECTED_NAME"
            printf "%b\n" "${CHECK} ${GREEN}Applied locally: $SELECTED_NAME <$SELECTED_EMAIL>${NC}"
        fi
    fi
fi
