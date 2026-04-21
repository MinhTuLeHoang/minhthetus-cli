#!/bin/bash
# Description: Automatically calculates the next version based on existing stg/qc tags, creates new annotated tags on the CURRENT branch, and pushes to origin.

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Git Tag Dev/Stg"
HELP_USAGE="minhthetus-cli git tag-dev-stg [options]"
HELP_DESCRIPTION="Automatically calculates the next version, creates stg-v* and qc-v* tags, and pushes to origin."
HELP_OPTIONS="-P, --patch        | Increment the patch version (e.g. 1.0.0 -> 1.0.1)\n-N, --minor        | Increment the minor version (e.g. 1.0.0 -> 1.1.0) [Default]\n-M, --major        | Increment the major version (e.g. 1.0.0 -> 2.0.0)\n-m <message>       | Provide a custom tag message"
HELP_EXAMPLE="minhthetus-cli git tag-dev-stg -P -m \"Hotfix for production\""

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

printf "%b\n" "${BLUE}${INFO} Starting Super Tag process...${NC}"


# Function to extract version from tag
get_version() {
    echo "$1" | sed 's/[^0-9.]*//g'
}

# Function to compare versions (x.y.z)
# Returns 1 if v1 > v2, 0 if v1 == v2, -1 if v1 < v2
compare_versions() {
    local v1=$1
    local v2=$2
    
    if [[ "$v1" == "$v2" ]]; then echo 0; return; fi
    
    local IFS=.
    local i t1 t2
    local -a v1_parts=($v1)
    local -a v2_parts=($v2)
    
    for ((i=0; i<3; i++)); do
        t1=${v1_parts[i]:-0}
        t2=${v2_parts[i]:-0}
        if ((t1 > t2)); then echo 1; return; fi
        if ((t1 < t2)); then echo -1; return; fi
    done
    echo 0
}

# 1. Search for latest git tags
printf "%b\n" "${CYAN}${INFO} Fetching latest tags from git...${NC}"
LATEST_STG_TAG=$(git tag -l "stg-v*" | sort -V | tail -n 1)
LATEST_QC_TAG=$(git tag -l "qc-v*" | sort -V | tail -n 1)

STG_VERSION=$(get_version "$LATEST_STG_TAG")
QC_VERSION=$(get_version "$LATEST_QC_TAG")

printf "\n"
printf "%b\n" "${TAG} Latest STG tag: ${YELLOW}${LATEST_STG_TAG:-"None"}${NC} (Version: ${STG_VERSION:-"0.0.0"})"
printf "%b\n" "${TAG} Latest QC tag:  ${YELLOW}${LATEST_QC_TAG:-"None"}${NC} (Version: ${QC_VERSION:-"0.0.0"})"

# Determine finalVersion (max)
if [[ -z "$STG_VERSION" ]] && [[ -z "$QC_VERSION" ]]; then
    FINAL_VERSION="0.0.0"
elif [[ -z "$STG_VERSION" ]]; then
    FINAL_VERSION="$QC_VERSION"
elif [[ -z "$QC_VERSION" ]]; then
    FINAL_VERSION="$STG_VERSION"
else
    CMP=$(compare_versions "$STG_VERSION" "$QC_VERSION")
    if [[ $CMP -ge 0 ]]; then
        FINAL_VERSION="$STG_VERSION"
    else
        FINAL_VERSION="$QC_VERSION"
    fi
fi

printf "%b\n" "${INFO} Base version identified: ${GREEN}${FINAL_VERSION}${NC}"

# 2. Parse flags for updating version
INCREMENT_TYPE="minor" # Default is -N
MESSAGE=""

# Create a temporary copy of arguments for parsing
ARGS=("$@")
while [[ ${#ARGS[@]} -gt 0 ]]; do
    case "${ARGS[0]}" in
        -P|--patch)
            INCREMENT_TYPE="patch"
            ARGS=("${ARGS[@]:1}")
            ;;
        -N|--minor)
            INCREMENT_TYPE="minor"
            ARGS=("${ARGS[@]:1}")
            ;;
        -M|--major)
            INCREMENT_TYPE="major"
            ARGS=("${ARGS[@]:1}")
            ;;
        -m)
            if [[ -n "${ARGS[1]}" && "${ARGS[1]}" != -* ]]; then
                MESSAGE="${ARGS[1]}"
                ARGS=("${ARGS[@]:2}")
            else
                # -m was provided without a message, we will prompt for it later
                ARGS=("${ARGS[@]:1}")
            fi
            ;;
        *)
            # Ignore unknown or handle error (since print-help handles -h/--help)
            ARGS=("${ARGS[@]:1}")
            ;;
    esac
done

# Increment logic
IFS=. read -r major minor patch <<< "$FINAL_VERSION"
major=${major:-0}
minor=${minor:-0}
patch=${patch:-0}

if [[ "$INCREMENT_TYPE" == "major" ]]; then
    major=$((major + 1))
    minor=0
    patch=0
elif [[ "$INCREMENT_TYPE" == "minor" ]]; then
    minor=$((minor + 1))
    patch=0
elif [[ "$INCREMENT_TYPE" == "patch" ]]; then
    patch=$((patch + 1))
fi

printf "\n"
NEW_VERSION="${major}.${minor}.${patch}"
printf "%b\n" "${HAMMER} Incrementing version (${INCREMENT_TYPE})..."
printf "%b\n" "${ROCKET} Final version to be used: ${PURPLE}${NEW_VERSION}${NC}"

# Handle tag message
if [[ -z "$MESSAGE" ]]; then
    if command -v gum &> /dev/null; then
        MESSAGE=$(gum input --placeholder "Enter tag message (e.g. Release v${NEW_VERSION})")
    else
        printf "\n"
        printf "%b\n" "${BLUE}${INFO} Please enter the tag message:${NC} "
        read -r MESSAGE
    fi
fi

if [[ -z "$MESSAGE" ]]; then
    MESSAGE="Release v${NEW_VERSION}"
fi

QC_TAG="qc-v${NEW_VERSION}"
STG_TAG="stg-v${NEW_VERSION}"

# Check current branch
printf "\n"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
printf "%b\n" "${INFO} Current branch: ${YELLOW}${CURRENT_BRANCH}${NC}"

# 3. Create tags on current branch
printf "\n"
printf "%b\n" "${CYAN}${INFO} Creating tags on ${YELLOW}${CURRENT_BRANCH}${NC} branch...${NC}"

# Ensure we are tagging the current branch
git tag -a "$QC_TAG" "$CURRENT_BRANCH" -m "$MESSAGE"
if [[ $? -eq 0 ]]; then
    printf "%b\n" "${CHECK} Created tag: ${GREEN}${QC_TAG}${NC}"
else
    printf "%b\n" "${RED}${ERROR} Failed to create tag ${QC_TAG}${NC}"
    exit 1
fi

git tag -a "$STG_TAG" "$CURRENT_BRANCH" -m "$MESSAGE"
if [[ $? -eq 0 ]]; then
    printf "%b\n" "${CHECK} Created tag: ${GREEN}${STG_TAG}${NC}"
else
    printf "%b\n" "${RED}${ERROR} Failed to create tag ${STG_TAG}${NC}"
    exit 1
fi

# 4. Push tags to origin
printf "\n"
printf "%b\n" "${CYAN}${INFO} Pushing tags to origin...${NC}"
git push origin "$QC_TAG" "$STG_TAG"

if [[ $? -eq 0 ]]; then
    printf "\n"
    printf "%b\n" "${GREEN}${CHECK} Successfully created and pushed both tags for version ${NEW_VERSION}!${NC}"
    printf "%b\n" "${BLUE}${TAG} QC Tag:  ${YELLOW}${QC_TAG}${NC}"
    printf "%b\n" "${BLUE}${TAG} STG Tag: ${YELLOW}${STG_TAG}${NC}"
else
    printf "%b\n" "${RED}${ERROR} Failed to push tags to origin.${NC}"
    exit 1
fi

