#!/bin/bash
# Description: Creates a backup of the current branch and maintains only 3 recent versions.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Git Backup Branch"
HELP_USAGE="minhthetus-cli git backup-branch [-l, --list]"
HELP_DESCRIPTION="Creates a backup branch named backup/<current-branch>-dd-mm-yyyy-HHh-MM.
Maintains up to 3 versions of backups for the current branch and prompts for cleanup if exceeded."
HELP_OPTIONS="-l, --list | List all backup branches for the current branch without creating a new one."
HELP_EXAMPLE="minhthetus-cli git backup-branch\nminhthetus-cli git backup-branch --list"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

set -euo pipefail

# Parse arguments
LIST_ONLY=false
for arg in "$@"; do
  case $arg in
    --list|-l)
      LIST_ONLY=true
      ;;
  esac
done

# Get current branch name
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
if [ -z "$CURRENT_BRANCH" ]; then
    printf "%b\n" "${ERROR} ${RED}Not a git repository or no branch found.${NC}"
    exit 1
fi

# Function to get matched backups with chronological sorting
get_matched_backups() {
    local branch=$1
    local all_backups_list=$(git branch --list "backup/${branch}-*" --format='%(refname:short)')
    local matched_with_keys=""
    
    while read -r b; do
        [ -z "$b" ] && continue
        # Regex: backup/branch-name-DD-MM-YYYY-HHh-MM
        if [[ "$b" =~ backup/(.*)-([0-9]{2})-([0-9]{2})-([0-9]{4})-([0-9]{2})h-([0-9]{2}) ]]; then
            # Ensure the branch name part matches exactly
            if [ "backup/${BASH_REMATCH[1]}" == "backup/${branch}" ]; then
                local dd="${BASH_REMATCH[2]}"
                local mm="${BASH_REMATCH[3]}"
                local yyyy="${BASH_REMATCH[4]}"
                local hh="${BASH_REMATCH[5]}"
                local min="${BASH_REMATCH[6]}"
                local key="${yyyy}${mm}${dd}${hh}${min}"
                matched_with_keys="${matched_with_keys}${key} ${b}\n"
            fi
        fi
    done <<< "$all_backups_list"
    
    printf "%b" "$matched_with_keys" | sort -rn | cut -d' ' -f2- | sed '/^$/d'
}

# Mode: List only
if [ "$LIST_ONLY" = true ]; then
    printf "\n"
    printf "%b\n" "${INFO} Listing all backups for branch: ${CYAN}${CURRENT_BRANCH}${NC}"
    
    MATCHED_BACKUPS=$(get_matched_backups "$CURRENT_BRANCH")
    
    if [ -z "$MATCHED_BACKUPS" ]; then
        printf "%b\n" "  ${INFO} No backups found for this branch."
    else
        printf "\n"
        while read -r b; do
            printf "  ${BULLET} ${GREEN}$b${NC}\n"
        done <<< "$MATCHED_BACKUPS"
    fi
    printf "\n"
    exit 0
fi

# Mode: Create backup
DATE_SUFFIX=$(date +%d-%m-%Y)
TIME_SUFFIX=$(date +%Hh-%M)
BACKUP_NAME="backup/${CURRENT_BRANCH}-${DATE_SUFFIX}-${TIME_SUFFIX}"

printf "\n"
printf "%b\n" "${INFO} Current branch: ${CYAN}${CURRENT_BRANCH}${NC}"

# Check if a backup with the same date/time already exists
if git show-ref --verify --quiet "refs/heads/${BACKUP_NAME}"; then
    printf "%b\n" "${ERROR} ${RED}Backup branch '${BACKUP_NAME}' already exists.${NC}"
    printf "%b\n" "${INFO} Please wait at least 1 minute before creating another backup.${NC}"
    exit 1
fi

printf "%b\n" "${HOURGLASS} Creating backup branch: ${GREEN}${BACKUP_NAME}${NC}..."

# Create the backup branch
git branch "$BACKUP_NAME"

# Push to origin
if git remote | grep -q "^origin$"; then
    git push origin "$BACKUP_NAME"
    printf "%b\n" "${CHECK} Backup created and pushed to origin: ${GREEN}${BACKUP_NAME}${NC}"
else
    printf "%b\n" "${CHECK} Backup created locally: ${GREEN}${BACKUP_NAME}${NC}"
fi

# Manage backup versions
MATCHED_BACKUPS=$(get_matched_backups "$CURRENT_BRANCH")

if [ -z "$MATCHED_BACKUPS" ]; then
    COUNT=0
else
    COUNT=$(echo "$MATCHED_BACKUPS" | wc -l | tr -d ' ')
fi

if [ "$COUNT" -gt 3 ]; then
    printf "\n"
    printf "%b\n" "${WARNING} Found ${YELLOW}${COUNT}${NC} backups for ${CYAN}${CURRENT_BRANCH}${NC}."
    printf "%b\n" "${INFO} Keeping the 3 latest versions. The following old backups can be deleted:"
    
    # The first 3 are the latest (due to sorting in get_matched_backups)
    OLD_BACKUPS=$(echo "$MATCHED_BACKUPS" | tail -n +4)
    
    printf "\n"
    for b in $OLD_BACKUPS; do
        printf "  ${RED}- $b${NC}\n"
    done
    printf "\n"
    
    # Use gum to confirm deletion
    GUM_BIN="$SCRIPT_DIR/../../../vendor/bin/gum"
    
    if [ -f "$GUM_BIN" ]; then
        if "$GUM_BIN" confirm "Do you want to delete these $(echo "$OLD_BACKUPS" | wc -l | tr -d ' ') old backups (locally and on origin)?"; then
            for b in $OLD_BACKUPS; do
                git branch -D "$b"
                if git remote | grep -q "^origin$"; then
                    git push origin --delete "$b" 2>/dev/null || true
                fi
                printf "%b\n" "${CHECK} Deleted: ${RED}$b${NC}"
            done
        else
            printf "%b\n" "${INFO} Deletion skipped."
        fi
    else
        printf "Do you want to delete these old backups (locally and on origin)? (y/N): "
        read -r CONFIRM
        if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
            for b in $OLD_BACKUPS; do
                git branch -D "$b"
                if git remote | grep -q "^origin$"; then
                    git push origin --delete "$b" 2>/dev/null || true
                fi
                printf "%b\n" "${CHECK} Deleted: ${RED}$b${NC}"
            done
        else
            printf "%b\n" "${INFO} Deletion skipped."
        fi
    fi
fi

printf "\n"
printf "%b\n" "${CHECK} ${GREEN}Backup workflow completed!${NC}"
