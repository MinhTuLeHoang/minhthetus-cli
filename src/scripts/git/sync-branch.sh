#!/bin/bash
# Description: Synchronizes the current branch's latest commit to dev and staging branches using rebase or cherry-pick logic.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Git Branch Sync"
HELP_USAGE="minhthetus-cli git sync-branch"
HELP_DESCRIPTION="Synchronizes the current branch's latest commit to dev and staging branches. Automatically handles rebase if the branch is linear or cherry-picks otherwise."
HELP_EXAMPLE="minhthetus-cli git sync-branch"

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

set -euo pipefail

# Get current branch and latest commit
ORIGINAL_BRANCH=$(git rev-parse --abbrev-ref HEAD)
LATEST_COMMIT=$(git rev-parse HEAD)
COMMIT_MESSAGE=$(git log -1 --pretty=%s)

printf "\n"
printf "%b\n" "${INFO} Starting sync script..."
printf "\n"
printf "%b\n" "${INFO} Original Branch: ${CYAN}$ORIGINAL_BRANCH${NC}"
printf "%b\n" "${INFO} Latest Commit: ${CYAN}$LATEST_COMMIT${NC}"
printf "%b\n" "${INFO} Commit Message: ${CYAN}$COMMIT_MESSAGE${NC}"
printf "\n"

# Ensure we have the latest info for dev and staging
git fetch origin dev
git fetch origin staging

# Use origin hashes for detection to avoid local branch drift issues
DEV_HASH=$(git rev-parse origin/dev 2>/dev/null || git rev-parse dev 2>/dev/null || echo "")
STAGING_HASH=$(git rev-parse origin/staging 2>/dev/null || git rev-parse staging 2>/dev/null || echo "")

printf "\n"
printf "%b\n" "${INFO} dev hash: ${BLUE}$DEV_HASH${NC}"
printf "%b\n" "${INFO} staging hash: ${BLUE}$STAGING_HASH${NC}"
printf "\n"

# Check if feature branch contains dev and staging
DEV_IS_ANCESTOR=false
if [ -n "$DEV_HASH" ] && git merge-base --is-ancestor "$DEV_HASH" "$LATEST_COMMIT"; then
    DEV_IS_ANCESTOR=true
fi

STAGING_IS_ANCESTOR=false
if [ -n "$STAGING_HASH" ] && git merge-base --is-ancestor "$STAGING_HASH" "$LATEST_COMMIT"; then
    STAGING_IS_ANCESTOR=true
fi

# Feature branch contains dev/staging if BOTH are ancestors.
# This allows 'pulling' all new commits from feature into dev/staging.
FEATURE_CONTAINS_DEV_STAGING=false
if [ "$DEV_IS_ANCESTOR" = true ] && [ "$STAGING_IS_ANCESTOR" = true ]; then
    FEATURE_CONTAINS_DEV_STAGING=true
fi

# Prioritize Linear Sync (Case A Rebase) if both branches are ancestors
if [ "$FEATURE_CONTAINS_DEV_STAGING" = true ]; then
    printf "%b\n" "${TAG} ${GREEN}Detected Linear Case: Both dev and staging are ancestors of the current branch.${NC}"
    printf "%b\n" "${TAG} ${GREEN}Syncing with rebase/FF (no cherry-pick needed).${NC}"
    printf "\n"

    # CASE A (rebase): checkout dev, pull --rebase, rebase onto feature, push; checkout staging, pull --rebase dev, push.
    printf "%b\n" "${HOURGLASS} Checking out dev..."
    git checkout dev

    printf "%b\n" "${HOURGLASS} Pulling latest dev from origin (rebase)..."
    git pull --rebase origin dev

    printf "%b\n" "${HAMMER} Rebasing dev onto $ORIGINAL_BRANCH ($LATEST_COMMIT)..."
    git rebase "$ORIGINAL_BRANCH" || {
        printf "%b\n" "${ERROR} ${RED}Rebase onto $ORIGINAL_BRANCH failed. Aborting rebase.${NC}"
        git rebase --abort 2>/dev/null || true
        git checkout "$ORIGINAL_BRANCH"
        printf "\n"
        printf "%b\n" "${INFO} FINAL STATUS:"
        printf "%b\n" "dev: ${ERROR} Rebase failed"
        printf "%b\n" "staging: ${INFO} Skipped"
        exit 1
    }

    git push origin dev --force
    printf "%b\n" "${CHECK} ${GREEN}dev updated and pushed.${NC}"
    printf "\n"

    printf "%b\n" "${HOURGLASS} Checking out staging..."
    git checkout staging

    printf "%b\n" "${HOURGLASS} Pulling dev into staging (rebase)..."
    git pull --rebase origin dev

    git push origin staging --force
    printf "%b\n" "${HAMMER} Sync dev to staging successfully"
    printf "\n"

    DEV_STATUS="${CHECK} Successfully rebased and pushed"
    STAGING_STATUS="${CHECK} Successfully pulled from dev (rebase) and pushed"

elif [ "$DEV_HASH" == "$STAGING_HASH" ]; then
    printf "%b\n" "${TAG} ${GREEN}Detected Case A: dev and staging are at the same node (not ancestors).${NC}"
    printf "\n"

    # CASE A (cherry-pick): checkout dev, cherry pick, push. checkout staging, pull dev, push.

        # 1. Checkout dev
        printf "%b\n" "${HOURGLASS} Checking out dev..."
        git checkout dev || { printf "%b\n" "${RED}${ERROR} Failed to checkout dev${NC}"; exit 1; }

        # 2. Cherry pick
        printf "%b\n" "${HAMMER} Cherry picking $LATEST_COMMIT into dev..."
        if git cherry-pick "$LATEST_COMMIT"; then
            # 3. Push dev
            git push origin dev || { printf "%b\n" "${RED}${ERROR} Failed to push dev${NC}"; exit 1; }
            printf "%b\n" "${CHECK} ${GREEN}Cherry pick into dev successful.${NC}"
            printf "\n"

            # 4. Checkout staging
            printf "%b\n" "${HOURGLASS} Checking out staging..."
            git checkout staging || { printf "%b\n" "${RED}${ERROR} Failed to checkout staging${NC}"; exit 1; }

            # 5. Pull dev into staging
            git pull --rebase origin dev || { printf "%b\n" "${RED}${ERROR} Failed to pull dev into staging${NC}"; exit 1; }

            # 6. Push staging
            git push origin staging || { printf "%b\n" "${RED}${ERROR} Failed to push staging${NC}"; exit 1; }
            printf "%b\n" "${HAMMER} Sync dev to staging successfully"
            printf "\n"

            DEV_STATUS="${CHECK} Successfully cherry-picked and pushed"
            STAGING_STATUS="${CHECK} Successfully pulled from dev and pushed"
        else
            printf "%b\n" "${ERROR} ${RED}Conflict occurred while cherry-picking into dev. Aborting Case A.${NC}"
            git cherry-pick --abort 2>/dev/null || true
            DEV_STATUS="${ERROR} Cherry-pick conflict"
            STAGING_STATUS="${INFO} Skipped"
            git checkout "$ORIGINAL_BRANCH"

            printf "\n"
            printf "%b\n" "${INFO} FINAL STATUS:"
            printf "%b\n" "dev: $DEV_STATUS"
            printf "%b\n" "staging: $STAGING_STATUS"
            exit 1
        fi
else
    printf "%b\n" "${TAG} ${YELLOW}Detected Case B: dev and staging are at different nodes.${NC}"
    printf "\n"
    
    # CASE B: checkout dev, cherry pick, push. checkout staging, cherry pick, push.
    
    # 1. Checkout dev and cherry pick
    printf "%b\n" "${HOURGLASS} Checking out dev..."
    git checkout dev || { printf "%b\n" "${RED}${ERROR} Failed to checkout dev${NC}"; exit 1; }
    
    printf "%b\n" "${HAMMER} Cherry picking $LATEST_COMMIT into dev..."
    if git cherry-pick "$LATEST_COMMIT"; then
        printf "%b\n" "${CHECK} ${GREEN}Cherry pick into dev successful.${NC}"
        printf "%b\n" "${ROCKET} Pushing dev..."
        git push origin dev || { printf "%b\n" "${RED}${ERROR} Failed to push dev${NC}"; exit 1; }
        DEV_STATUS="${CHECK} Successfully cherry-picked and pushed"
    else
        printf "%b\n" "${ERROR} ${RED}Conflict occurred while cherry-picking into dev. Aborting dev update.${NC}"
        git cherry-pick --abort 2>/dev/null || true
        DEV_STATUS="${ERROR} Cherry-pick conflict"
        STAGING_STATUS="${INFO} Skipped due to dev failure"
        git checkout "$ORIGINAL_BRANCH"
        
        printf "\n"
        printf "%b\n" "${INFO} FINAL STATUS:"
        printf "%b\n" "dev: $DEV_STATUS"
        printf "%b\n" "staging: $STAGING_STATUS"
        exit 1
    fi
    
    # 2. Checkout staging and cherry pick
    printf "%b\n" "${HOURGLASS} Checking out staging..."
    git checkout staging || { printf "%b\n" "${RED}${ERROR} Failed to checkout staging${NC}"; exit 1; }
    
    printf "%b\n" "${HAMMER} Cherry picking $LATEST_COMMIT into staging..."
    if git cherry-pick "$LATEST_COMMIT"; then
        printf "%b\n" "${CHECK} ${GREEN}Cherry pick into staging successful.${NC}"
        printf "%b\n" "${ROCKET} Pushing staging..."
        git push origin staging || { printf "%b\n" "${RED}${ERROR} Failed to push staging${NC}"; exit 1; }
        STAGING_STATUS="${CHECK} Successfully cherry-picked and pushed"
    else
        printf "%b\n" "${ERROR} ${RED}Conflict occurred while cherry-picking into staging. Aborting staging update.${NC}"
        git cherry-pick --abort 2>/dev/null || true
        STAGING_STATUS="${ERROR} Cherry-pick conflict"
        git checkout "$ORIGINAL_BRANCH"
        printf "\n"
        printf "%b\n" "${INFO} FINAL STATUS:"
        printf "%b\n" "dev: $DEV_STATUS"
        printf "%b\n" "staging: $STAGING_STATUS"
        exit 1
    fi
fi

# Return to original branch
printf "\n"
printf "%b\n" "${HOURGLASS} Returning to original branch: $ORIGINAL_BRANCH"
git checkout "$ORIGINAL_BRANCH"

# Inform final status
printf "\n"
printf "%b\n" "${INFO} FINAL STATUS:"
printf "%b\n" "dev: $DEV_STATUS"
printf "%b\n" "staging: $STAGING_STATUS"

if [[ "$DEV_STATUS" == *"${CHECK}"* && "$STAGING_STATUS" == *"${CHECK}"* ]]; then
    printf "%b\n" "${CHECK} ${GREEN}All branches updated successfully!${NC}"
else
    printf "%b\n" "${ERROR} ${RED}Some updates failed or were skipped.${NC}"
    exit 1
fi

