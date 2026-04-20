#!/bin/bash
# Description: Recursively applies chmod +x to all .sh files within the src directory.

# Get the directories
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SRC_DIR="$( dirname "$SCRIPT_DIR" )"

# Source constants
source "$SCRIPT_DIR/constants.sh"

echo -e "${BLUE}${ROCKET} Applying chmod +x to all .sh files in ${YELLOW}$SRC_DIR${NC}..."

# Find all .sh files in the src directory and make them executable
find "$SRC_DIR" -name "*.sh" -exec chmod +x {} \;

echo -e "${GREEN}${CHECK} All .sh files in src are now executable!${NC}"
