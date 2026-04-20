#!/bin/bash
# Description: Demonstrates gum enhanced UI interactions

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
ACTION=$(gum choose "Say Hello" "System Info" "Spin Demo" "Exit")

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
    "Exit")
        gum style --foreground 240 "Goodbye!"
        exit 0
        ;;
esac
