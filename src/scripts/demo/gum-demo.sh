#!/bin/bash
# Description: Demonstrates gum enhanced UI interactions

# Source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GENERAL_SCRIPTS_DIR="$SCRIPT_DIR/../../generalScripts"

HELP_TITLE="Gum Demo"
HELP_USAGE="minhthetus-cli demo gum-demo"
HELP_DESCRIPTION="An interactive demonstration of gum-enhanced UI elements like choose, input, spin, and confirm."

source "$GENERAL_SCRIPTS_DIR/print-help.sh" "$@"

# Check if gum is in PATH
if ! command -v gum &> /dev/null; then
    echo "Error: gum is not installed or not in PATH."
    exit 1
fi

# 1. Stylized Header
gum style \
	--foreground 212 --border-foreground 212 --border double \
	--align center --width 50 --margin "1 2" --padding "1 2" \
	"GUM INTERACTION DEMO"

# 2. Interactive Choice
ACTION=$(gum choose "Say Hello" "System Info" "Spin Demo" "Confirm Demo" "Gum Version" "Exit")

case "$ACTION" in
    "Say Hello")
        NAME=$(gum input --placeholder "What is your name?")
        gum style --foreground 86 "Hello, $NAME! Welcome to the enhanced CLI."
        ;;
    "System Info")
        gum style --border normal --margin "1 2" --padding "1 2" --border-foreground 99 \
            "OS: $(uname -s)" \
            "Kernel: $(uname -r)" \
            "Shell: $SHELL"
        ;;
    "Spin Demo")
        gum spin --spinner dot --title "Performing background magic..." -- sleep 2
        gum style --foreground 2 green "✔ Magic complete!"
        ;;
    "Confirm Demo")
        # Display auto-approve guide
        gum style --foreground 245 "(Auto-approving in 2 seconds...)"
        
        if gum confirm "Proceed with the operation?" --timeout=2s --default="Yes"; then
            gum style --foreground 2 --bold "Status: CONFIRMED (or auto-approved)"
        else
            gum style --foreground 1 --bold "Status: CANCELLED (Exit: $?)"
        fi
        ;;
    "Gum Version")
        gum --version
        ;;
    "Exit")
        gum style --foreground 240 "Goodbye!"
        exit 0
        ;;
esac
