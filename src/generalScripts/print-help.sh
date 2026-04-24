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
#   HELP_TAB_SIZE=4  # Optional: change left indentation (default: 3)
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

# Configuration
_GENERAL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[ -f "$_GENERAL_DIR/constants.sh" ] && source "$_GENERAL_DIR/constants.sh"

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

    printf "\n"
    printf "%b%s\n" "${BLUE}${INFO} ${BOLD}${title} Usage Guide:${NC}" ""
    printf "%s%s\n" "${HELP_INDENT}" "${usage}"
    printf "\n"
    printf "%b%s\n" "${YELLOW}${BOLD}Description:${NC}" ""
    printf "%b\n" "$description" | while IFS= read -r line; do
        [ -n "$line" ] && printf "%s%s\n" "${HELP_INDENT}" "${line}"
    done
    printf "%b\n" ""
    printf "%b\n" "${YELLOW}${BOLD}Options:${NC}"
    if [[ "$options" == *"|"* ]]; then
        printf "%b\n" "$options" | while IFS='|' read -r col1 col2; do
            if [[ -n "$col2" ]]; then
                # Clean whitespace and print with fixed width for column 1
                col1_clean=$(echo "$col1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
                col2_clean=$(echo "$col2" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
                printf "%b%-22s%b %s\n" "${HELP_INDENT}${CYAN}" "$col1_clean" "${NC}" "$col2_clean"
            else
                col1_clean=$(echo "$col1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
                [ -n "$col1_clean" ] && printf "%s%s\n" "${HELP_INDENT}" "$col1_clean"
            fi
        done
    else
        printf "%b\n" "$options" | while IFS= read -r line; do
             [ -n "$line" ] && printf "%s%s\n" "${HELP_INDENT}" "${line}"
        done
    fi
    printf "%b\n" ""
    printf "%b\n" "${YELLOW}${BOLD}Example:${NC}"
    printf "%b\n" "$example" | while IFS= read -r line; do
        [ -n "$line" ] && printf "%s%s\n" "${HELP_INDENT}" "${line}"
    done
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
            # Automatically add help option if not already present
            if [[ "$HELP_OPTIONS" != *"--help"* ]]; then
                if [[ -n "$HELP_OPTIONS" ]]; then
                    HELP_OPTIONS="${HELP_OPTIONS}\n-h, --help       | Show this help message."
                else
                    HELP_OPTIONS="-h, --help       | Show this help message."
                fi
            fi
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