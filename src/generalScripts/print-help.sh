#!/bin/bash
# Description: Dynamic help message handler and renderer for CLI scripts.

# This script provides a standardized way to display help messages with consistent styling.
# It can be sourced by other scripts to automatically handle -h and --help flags.
#
# How to use:
# -----------------------------------------------------------------------------
# Method A: Use HELP_* variables (Recommended)
# Define these variables BEFORE sourcing this script.
# Use a pipe '|' to automatically align options and descriptions.
#
#   HELP_TITLE="My Script"
#   HELP_USAGE="./my-script.sh [options]"
#   HELP_DESCRIPTION="What the script does."
#   HELP_OPTIONS="--opt      | Description\n--flag <v> | Description"
#   HELP_EXAMPLE="./my-script.sh --opt"
#   source "$(dirname "$0")/print-help.sh" "$@"
#
# Method B: Custom print_usage() function
# Define this function BEFORE sourcing this script.
#
#   print_usage() {
#     echo "${BLUE}${INFO} Custom Help:${NC}"
#     echo "  ./my-script.sh [args]"
#   }
#   source "$(dirname "$0")/print-help.sh" "$@"
# -----------------------------------------------------------------------------


# Colors (fallback if constants.sh is not sourced)
[ -z "$NC" ] && NC='\033[0m'
[ -z "$BLUE" ] && BLUE='\033[0;34m'
[ -z "$YELLOW" ] && YELLOW='\033[0;33m'
[ -z "$GREEN" ] && GREEN='\033[0;32m'
[ -z "$PURPLE" ] && PURPLE='\033[0;35m'
[ -z "$CYAN" ] && CYAN='\033[0;36m'
[ -z "$RED" ] && RED='\033[0;31m'

# Icons (fallback if constants.sh is not sourced)
[ -z "$INFO" ] && INFO="ℹ️"
[ -z "$ERROR" ] && ERROR="❌"

# Function to render a formatted help message
# Arguments:
# 1: Title (e.g., "Super Tag Script")
# 2: Usage (e.g., "./super-tag.sh [options]")
# 3: Description
# 4: Options (formatted string)
# 5: Example
render_help() {
    local title="$1"
    local usage="$2"
    local description="$3"
    local options="$4"
    local example="$5"

    printf "%b\n" ""
    printf "%b\n" "${BLUE}${INFO} ${title} Usage Guide:${NC}"
    printf "%b\n" "  ${usage}"
    printf "%b\n" ""
    printf "%b\n" "${YELLOW}Description:${NC}"
    printf "%b\n" "  ${description}"
    printf "%b\n" ""
    printf "%b\n" "${YELLOW}Options:${NC}"
    if [[ "$options" == *"|"* ]]; then
        printf "%b\n" "$options" | while IFS='|' read -r col1 col2; do
            if [[ -n "$col2" ]]; then
                # Clean whitespace and print with fixed width for column 1
                col1_clean=$(echo "$col1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
                col2_clean=$(echo "$col2" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
                printf "  ${CYAN}%-22s${NC} %s\n" "$col1_clean" "$col2_clean"
            else
                printf "%b\n" "  $col1"
            fi
        done
    else
        printf "%b\n" "${options}"
    fi
    printf "%b\n" ""
    printf "%b\n" "${YELLOW}Example:${NC}"
    printf "%b\n" "  ${example}"
    printf "%b\n" ""
}

# Main logic when sourced or executed
# It checks if -h or --help is present in the arguments.
# If so, it looks for HELP_* variables and displays the help.

check_and_print_help() {
    local found_help=false
    for arg in "$@"; do
        if [[ "$arg" == "-h" || "$arg" == "--help" ]]; then
            found_help=true
            break
        fi
    done

    if [[ "$found_help" == "true" ]]; then
        # If the caller defined HELP_* variables, use them to render
        if [[ -n "$HELP_TITLE" ]]; then
            render_help "$HELP_TITLE" "$HELP_USAGE" "$HELP_DESCRIPTION" "$HELP_OPTIONS" "$HELP_EXAMPLE"
            exit 0
        # Alternatively, if they defined a print_usage function
        elif declare -f print_usage > /dev/null; then
            print_usage
            exit 0
        else
            printf "%b\n" "${RED}${ERROR} Error: Help requested but no HELP_* variables or print_usage() function defined.${NC}"
            exit 1
        fi
    fi
}

# Execute check immediately with passed arguments
check_and_print_help "$@"