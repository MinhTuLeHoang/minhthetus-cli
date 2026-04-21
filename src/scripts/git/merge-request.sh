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
printf "%b\n" "${BLUE}${ROCKET} Committing changes...${NC}"
git add .
if git diff --staged --quiet; then
    printf "%b\n" "${YELLOW}${INFO} No changes to commit.${NC}"
else
    git commit -m "$COMMIT_MESSAGE"
fi

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
        gh pr create --base master --head "$CURRENT_BRANCH" --title "$COMMIT_MESSAGE" --body "Automatically created by minhthetus-cli"
        SUCCESS=true
    fi

# Case 2: GitLab (using push options)
elif [[ "$REMOTE_URL" == *"gitlab"* ]]; then
    printf "%b\n" "${CYAN}${INFO} Utilizing GitLab push options for MR creation...${NC}"
    if git push -o mr.create -o mr.target=master -o mr.title="$COMMIT_MESSAGE" -o mr.remove_source_branch origin "$CURRENT_BRANCH"; then
        SUCCESS=true
    fi

# Case 3: Agit Fallback (Gitea, etc.)
else
    printf "%b\n" "${CYAN}${INFO} Attempting to create PR via Agit (refs/for/master)...${NC}"
    # Agit requires a topic (-o topic). We use the current branch name.
    if git push origin "HEAD:refs/for/master" -o topic="$CURRENT_BRANCH" -o title="$COMMIT_MESSAGE" -o description="Automatically created by minhthetus-cli"; then
        SUCCESS=true
    else
        printf "%b\n" "${YELLOW}${WARNING} Agit push failed. This host might not support Agit.${NC}"
    fi
fi

# Fallback: Print Manual Link if automation was not possible or failed
if [ "$SUCCESS" = false ]; then
    printf "\n"
    printf "%b\n" "${YELLOW}${WARNING} Could not automate PR/MR creation. Manual Link:${NC}"
    if [[ "$REMOTE_URL" == *"github.com"* ]]; then
        # Extract user/repo from formats like git@github.com:user/repo.git or https://github.com/user/repo.git
        REPO_PATH=$(echo "$REMOTE_URL" | sed -E 's/.*github.com[:\/](.*)\.git/\1/')
        printf "  https://github.com/$REPO_PATH/compare/master...$CURRENT_BRANCH?expand=1\n"
    elif [[ "$REMOTE_URL" == *"gitlab"* ]]; then
        # Best-effort extraction for GitLab
        REPO_PATH=$(echo "$REMOTE_URL" | sed -E 's/.*gitlab.*[:\/](.*)\.git/\1/')
        # Extract domain (e.g. gitlab.com or self-hosted)
        DOMAIN=$(echo "$REMOTE_URL" | sed -E 's/.*@([^:\/]+).*/\1/' | sed -E 's/http(s)?:\/\///')
        printf "  https://$DOMAIN/$REPO_PATH/-/merge_requests/new?merge_request[source_branch]=$CURRENT_BRANCH&merge_request[target_branch]=master\n"
    else
        printf "  Please visit your git provider's web interface to open a PR/MR manually.\n"
    fi
fi

printf "\n"
printf "%b\n" "${GREEN}${CHECK} Workflow complete!${NC}"
