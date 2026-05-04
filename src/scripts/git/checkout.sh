#!/bin/bash
# Description: Checkout existing branches or create new ones using JIRA ticket IDs.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Git Checkout (JIRA)"
HELP_USAGE="minhthetus-cli git checkout [-j, --jira-ticket <id>]"
HELP_DESCRIPTION="Checks out an existing branch matching a JIRA ID, or creates a new one following a standardized naming convention. Supports numeric shorthand (e.g. 9404 -> PCFBANK-9404)."
HELP_OPTIONS="-j, --jira-ticket | JIRA ticket ID or number. If omitted, the script will prompt for it."
HELP_EXAMPLE="minhthetus-cli git checkout --jira-ticket PCFBANK-9404\nminhthetus-cli git checkout"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

# Exit on error, undefined variables, and pipe failures
set -euo pipefail

# Parse arguments
JIRA_ID=""
while [[ $# -gt 0 ]]; do
  case $1 in
    --jira-ticket|-j)
      JIRA_ID="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Interactive JIRA prompt if not provided
if [ -z "$JIRA_ID" ]; then
    JIRA_ID=$(gum input --placeholder "Ex: 9404 or PCFBANK-9404, type 'none' or empty to ignore")
    printf "\n"
fi

# Normalize JIRA_ID
if [[ "$JIRA_ID" == "none" ]] || [[ -z "$JIRA_ID" ]]; then
    JIRA_ID=""
else
    # Auto-prefix if input is only numbers
    if [[ "$JIRA_ID" =~ ^[0-9]+$ ]]; then
        JIRA_ID="PCFBANK-$JIRA_ID"
    fi
    # Upper case JIRA ID for consistency
    JIRA_ID=$(echo "$JIRA_ID" | tr '[:lower:]' '[:upper:]')
fi

# --- Discovery Logic ---
MATCHED_BRANCHES=""
if [ -n "$JIRA_ID" ]; then
    printf "%b\n" "${INFO} Searching for branches matching: ${CYAN}${JIRA_ID}${NC}"
    # Search locally
    MATCHED_BRANCHES=$(git branch --list "*${JIRA_ID}*" --format='%(refname:short)')
fi

if [ -n "$MATCHED_BRANCHES" ]; then
    COUNT=$(echo "$MATCHED_BRANCHES" | wc -l | tr -d ' ')
    if [ "$COUNT" -eq 1 ]; then
        SELECTED_BRANCH=$(echo "$MATCHED_BRANCHES" | xargs)
        printf "%b\n" "${CHECK} Found matching branch: ${GREEN}${SELECTED_BRANCH}${NC}"
        if git checkout "$SELECTED_BRANCH"; then
            printf "\n"
            printf "%b\n" "${HOURGLASS} Fetching latest updates from origin..."
            git pull origin "$SELECTED_BRANCH" 2>/dev/null || printf "%b\n" "${WARNING} ${YELLOW}Could not pull from origin. Remote branch might not exist.${NC}"
            printf "\n"
            exit 0
        else
            exit 1
        fi
    else
        printf "%b\n" "${INFO} Multiple matching branches found. Please select one:"
        SELECTED_BRANCH=$(echo "$MATCHED_BRANCHES" | gum choose)
        if [ -n "$SELECTED_BRANCH" ]; then
            printf "\n"
            if git checkout "$SELECTED_BRANCH"; then
                printf "\n"
                printf "%b\n" "${HOURGLASS} Fetching latest updates from origin..."
                git pull origin "$SELECTED_BRANCH" 2>/dev/null || printf "%b\n" "${WARNING} ${YELLOW}Could not pull from origin. Remote branch might not exist.${NC}"
                printf "\n"
                exit 0
            else
                exit 1
            fi
        fi
    fi
fi

# --- Creation Logic ---
printf "\n"
printf "%b\n" "${INFO} No existing branch found. Entering creation flow..."
printf "\n"

# 1. Select type
TYPES=("feature" "hotfix" "test" "docs" "improve" "bugfix" "refactor")
printf "%b\n" "${INFO} Select branch type:"
TYPE=$(gum choose "${TYPES[@]}")
# Clear the "Select branch type:" line and show result
printf "\033[1A\033[K%b\n" "${INFO} Branch type: ${CYAN}${TYPE}${NC}"

# 2. Input description
printf "%b\n" "${INFO} Enter branch name/description:"
DESC=$(gum input --placeholder "e.g. update user profile")

if [ -z "$DESC" ]; then
    printf "%b\n" "${ERROR} ${RED}Branch name/description cannot be empty.${NC}"
    exit 1
fi
# Clear the "Enter branch name/description:" line and show result
printf "\033[1A\033[K%b\n" "${INFO} Description: ${CYAN}${DESC}${NC}"

# 3. Format description
# Lowercase, replace spaces with dashes, remove special characters
FORMATTED_DESC=$(echo "$DESC" | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g' | sed 's/[^a-z0-9-]//g' | sed 's/--\+/-/g')

# 4. Construct final name
if [ -n "$JIRA_ID" ]; then
    FINAL_NAME="${TYPE}/${JIRA_ID}-${FORMATTED_DESC}"
else
    FINAL_NAME="${TYPE}/${FORMATTED_DESC}"
fi

printf "%b\n" "${HOURGLASS} Creating and checking out: ${GREEN}${FINAL_NAME}${NC}..."

# 5. Checkout
printf "\n"
if git checkout -b "$FINAL_NAME"; then
    printf "%b\n" "${CHECK} ${GREEN}Successfully created and checked out ${FINAL_NAME}${NC}"
    
    # 6. Push to origin
    printf "%b\n" "${HOURGLASS} Pushing to origin..."
    if git push -u origin "$FINAL_NAME"; then
        printf "%b\n" "${CHECK} ${GREEN}Successfully pushed to origin and set up tracking.${NC}"
    else
        printf "%b\n" "${WARNING} ${YELLOW}Failed to push to origin. You may need to push manually.${NC}"
    fi
    printf "\n"
else
    printf "%b\n" "${ERROR} ${RED}Failed to create branch.${NC}"
    printf "\n"
    exit 1
fi
