#!/bin/bash
# Description: Automatically bumps version, commits changes, and creates a Merge/Pull Request to the master branch.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Git Merge Request"
HELP_USAGE="minhthetus-cli git merge-request [options]"
HELP_DESCRIPTION="Automatically detects bump type from branch prefix, updates version, and opens a MR/PR to master."
HELP_OPTIONS="-M, --major      | Force major version bump\n-N, --minor      | Force minor version bump\n-P, --patch      | Force patch version bump\n--no-version     | Skip version bump step\n-m <message>     | Commit message (default: [bump version])"
HELP_EXAMPLE="minhthetus-cli git merge-request -m \"Add user authentication\""

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

set -e

# --- 1. Detect bump type and commit message ---
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
INCREMENT_TYPE=""
COMMIT_MESSAGE=""
SKIP_VERSION=false

# Parse arguments
ARGS=("$@")
while [[ ${#ARGS[@]} -gt 0 ]]; do
    case "${ARGS[0]}" in
        -M|--major)
            INCREMENT_TYPE="major"
            ARGS=("${ARGS[@]:1}")
            ;;
        -N|--minor)
            INCREMENT_TYPE="minor"
            ARGS=("${ARGS[@]:1}")
            ;;
        -P|--patch)
            INCREMENT_TYPE="patch"
            ARGS=("${ARGS[@]:1}")
            ;;
        -m)
            if [[ -n "${ARGS[1]}" && "${ARGS[1]}" != -* ]]; then
                COMMIT_MESSAGE="${ARGS[1]}"
                ARGS=("${ARGS[@]:2}")
            else
                COMMIT_MESSAGE="_PROMPT_"
                ARGS=("${ARGS[@]:1}")
            fi
            ;;
        --no-version)
            SKIP_VERSION=true
            ARGS=("${ARGS[@]:1}")
            ;;
        *)
            ARGS=("${ARGS[@]:1}")
            ;;
    esac
done

if [[ -z "$INCREMENT_TYPE" ]]; then
    if [[ $CURRENT_BRANCH =~ ^(fix/|hotfix/|docs/|test/|debug/) ]]; then
        INCREMENT_TYPE="patch"
    else
        INCREMENT_TYPE="minor"
    fi
fi

if [[ "$COMMIT_MESSAGE" == "_PROMPT_" ]]; then
    if command -v gum &> /dev/null; then
        COMMIT_MESSAGE=$(gum input --placeholder "Enter commit message (default: [bump version])")
    else
        printf "%b" "${BLUE}${INFO} Enter commit message: ${NC}"
        read -r COMMIT_MESSAGE
    fi
fi

if [[ -z "$COMMIT_MESSAGE" ]]; then
    COMMIT_MESSAGE="[bump version]"
fi

# --- 1.5. Sync with master and prepare MR content ---
printf "%b\n" "${BLUE}${HOURGLASS} Syncing with master before workflow...${NC}"
git fetch origin master

# Check if there is anything to rebase (avoid unnecessary rebase)
if ! git rebase origin/master; then
    printf "%b\n" "${RED}${ERROR} Rebase conflict detected. Please resolve manually and try again.${NC}"
    git rebase --abort
    exit 1
fi

# MR/PR Formatting Logic
JIRA_TICKET=$(echo "$CURRENT_BRANCH" | grep -oE "[A-Z]+-[0-9]+" | head -1) || JIRA_TICKET=""
PREFIX=$(echo "$CURRENT_BRANCH" | grep -oE "^[^/]+/" | sed 's/\///') || PREFIX=""
PRETTY_PREFIX=""
if [[ -n "$PREFIX" ]]; then
    PRETTY_PREFIX="$(tr '[:lower:]' '[:upper:]' <<< ${PREFIX:0:1})${PREFIX:1}/ "
fi

if [[ -n "$JIRA_TICKET" ]]; then
    # Extract rest of branch name by removing prefix/ticket and replacing separators with space
    REST_OF_BRANCH=$(echo "$CURRENT_BRANCH" | sed -E "s|^([^/]+/)?$JIRA_TICKET||; s|^[^/]+/||; s/^[-_]//; s/[-_]/ /g; s/  / /g")
    MR_TITLE="Resolve $JIRA_TICKET \"$PRETTY_PREFIX$REST_OF_BRANCH\""
    MR_TITLE=$(echo "$MR_TITLE" | sed 's/ " / "/') # Clean up space if any
    CLOSES_TEXT="Closes $JIRA_TICKET"
else
    MR_TITLE="$COMMIT_MESSAGE"
    CLOSES_TEXT=""
fi

# Generate commit list (diff between master and current branch)
COMMIT_LIST=$(git log origin/master..HEAD --oneline --format="- %s")
MR_DESCRIPTION=$(printf "%s\n\n%s" "$CLOSES_TEXT" "$COMMIT_LIST")

# --- 2. Bump version ---
if [[ "$SKIP_VERSION" == "true" ]]; then
    printf "%b\n" "${BLUE}${INFO} Skipping version bump as requested.${NC}"
elif [ -f "package.json" ]; then
    printf "%b\n" "${BLUE}${HAMMER} Bumping ${INCREMENT_TYPE} version...${NC}"
    
    npm version "$INCREMENT_TYPE" --no-git-tag-version
else
    printf "%b\n" "${YELLOW}${WARNING} No package.json found. Skipping version bump.${NC}"
fi

# --- 3. Stage, Commit, and Push ---
printf "\n"
printf "%b\n" "${BLUE}${ROCKET} Committing changes...${NC}"
git add .
if git diff --staged --quiet; then
    printf "%b\n" "${YELLOW}${INFO} No changes to commit.${NC}"
else
    git commit -m "$COMMIT_MESSAGE"
fi

printf "\n"
printf "%b\n" "${BLUE}${ROCKET} Pushing current branch to origin...${NC}"
git push origin "$CURRENT_BRANCH"

# --- 4. Create Merge Request / Pull Request ---
REMOTE_URL=$(git remote get-url origin)
printf "\n"
printf "%b\n" "${BLUE}${HOURGLASS} Orchestrating Merge Request/Pull Request to master...${NC}"

# Track if we successfully automated the creation
SUCCESS=false

# Case 1: GitHub
if [[ "$REMOTE_URL" == *"github.com"* ]]; then
    if command -v gh &> /dev/null; then
        gh pr create --base master --head "$CURRENT_BRANCH" --title "$MR_TITLE" --body "$MR_DESCRIPTION"
        SUCCESS=true
    fi

# Case 2: GitLab (using push options)
elif [[ "$REMOTE_URL" == *"gitlab"* ]]; then
    printf "%b\n" "${CYAN}${INFO} Utilizing GitLab push options for MR creation...${NC}"
    if git push -o mr.create -o mr.target=master -o mr.title="$MR_TITLE" -o mr.description="$MR_DESCRIPTION" -o mr.remove_source_branch origin "$CURRENT_BRANCH"; then
        SUCCESS=true
    fi

# Case 3: Agit Fallback (Gitea, etc.)
else
    printf "%b\n" "${CYAN}${INFO} Attempting to create PR via Agit (refs/for/master)...${NC}"
    # Agit requires a topic (-o topic). We use the current branch name.
    if git push origin "HEAD:refs/for/master" -o topic="$CURRENT_BRANCH" -o title="$MR_TITLE" -o description="$MR_DESCRIPTION"; then
        SUCCESS=true
    else
        printf "%b\n" "${YELLOW}${WARNING} Agit push failed. This host might not support Agit.${NC}"
    fi
fi

# --- 5. Manual Link Generation (ALWAYS shown for GitHub and GitLab) ---
printf "\n"
if [ "$SUCCESS" = true ]; then
    printf "%b\n" "${GREEN}${CHECK} Automation successful! For reference, your manual link is:${NC}"
else
    printf "%b\n" "${YELLOW}${WARNING} Could not automate PR/MR creation. Manual link:${NC}"
fi

if [[ "$REMOTE_URL" == *"github.com"* ]]; then
    # Extract user/repo from formats like git@github.com:user/repo.git or https://github.com/user/repo.git
    REPO_PATH=$(echo "$REMOTE_URL" | sed -E 's/.*github.com[:\/](.*)\.git/\1/')
    
    # URL Encode Title and Body for the manual link
    TITLE_ENC=$(echo "$MR_TITLE" | sed 's/ /%20/g; s/\[/%5B/g; s/\]/%5D/g; s/"/%22/g')
    BODY_ENC=$(echo "$MR_DESCRIPTION" | sed 's/ /%20/g; s/\[/%5B/g; s/\]/%5D/g; s/"/%22/g; s/$/%0A/g' | tr -d '\n')
    
    printf "  %s\n" "https://github.com/$REPO_PATH/compare/master...$CURRENT_BRANCH?expand=1&quick_pull=1&title=$TITLE_ENC&body=$BODY_ENC"
elif [[ "$REMOTE_URL" == *"gitlab"* ]]; then
    # Improved extraction for GitLab (handles subgroups correctly)
    REPO_PATH=$(echo "$REMOTE_URL" | sed -E 's/.*@//;s/http(s)?:\/\///;s/^[^\/:]+[\/:]//;s/\.git$//')
    DOMAIN=$(echo "$REMOTE_URL" | sed -E 's/.*@//;s/http(s)?:\/\///;s/[:\/].*//')
    
    # URL Encode Title and Description for the manual link
    TITLE_ENC=$(echo "$MR_TITLE" | sed 's/ /%20/g; s/\[/%5B/g; s/\]/%5D/g; s/"/%22/g')
    DESC_ENC=$(echo "$MR_DESCRIPTION" | sed 's/ /%20/g; s/\[/%5B/g; s/\]/%5D/g; s/"/%22/g; s/$/%0A/g' | tr -d '\n')
    
    printf "  %s\n" "https://$DOMAIN/$REPO_PATH/-/merge_requests/new?merge_request[source_branch]=$CURRENT_BRANCH&merge_request[target_branch]=master&merge_request[title]=$TITLE_ENC&merge_request[description]=$DESC_ENC"
elif [ "$SUCCESS" = false ]; then
    printf "  Please visit your git provider's web interface to open a PR/MR manually.\n"
fi

printf "\n"
printf "%b\n" "${GREEN}${CHECK} Workflow complete!${NC}"
