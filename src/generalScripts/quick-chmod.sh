#!/bin/bash
# Description: Recursively applies chmod +x to all .sh files within the src directory.

# Get the directories
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC_DIR="$( dirname "$SCRIPT_DIR" )"

printf "%b\n" "${BLUE}${ROCKET} Applying chmod +x to all .sh files in ${YELLOW}$SRC_DIR${NC}..."

# Find all .sh files in the src directory and make them executable
find "$SRC_DIR" -name "*.sh" -exec chmod +x {} \;

printf "%b\n" "${GREEN}${CHECK} All .sh files in src are now executable!${NC}"
