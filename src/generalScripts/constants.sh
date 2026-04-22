#!/bin/bash
# Description: Global constants, colors, and utility functions for shell scripts.

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Icons
CHECK="✅"
ERROR="❌"
INFO="ℹ️ "
WARNING="⚠️"
TAG="🏷️"

ROCKET="🚀"
HAMMER="🔨"
HOURGLASS="⏳"

# Spinner function
show_spinner() {
    local pid=$1
    local msg="$2"
    local log_file="$3"
    local delay=0.1
    local spinstr=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")
    local spin_count=${#spinstr[@]}
    local count=0
    local last_pos=1
    
    # Get high-res start time (milliseconds) using ruby
    local start_ms=$(ruby -e 'puts (Time.now.to_f * 1000).to_i' 2>/dev/null || date +%s000)
    
    # Hide cursor
    printf "\033[?25l"
    
    while kill -0 "$pid" 2>/dev/null; do
        # Show logs if log file is provided
        if [ -n "$log_file" ] && [ -f "$log_file" ]; then
            local new_logs=$(tail -c +$last_pos "$log_file")
            if [ -n "$new_logs" ]; then
                # Clear spinner line and print logs
                printf "\r\033[K%s" "$new_logs"
                # Update position for next read (byte count)
                local log_bytes=$(echo -n "$new_logs" | wc -c | tr -d '[:space:]')
                last_pos=$((last_pos + log_bytes))
            fi
        fi

        local now_ms=$(ruby -e 'puts (Time.now.to_f * 1000).to_i' 2>/dev/null || date +%s000)
        local elapsed=$((now_ms - start_ms))
        local sec=$((elapsed / 1000))
        local ms=$(((elapsed % 1000) / 100))
        local char="${spinstr[$((count % spin_count))]}"
        
        printf "\r${_INDENT:-  }${msg} %b%s%b (${sec}.${ms}s)  " "${CYAN}" "$char" "${NC}"
        sleep $delay
        count=$((count + 1))
    done
    
    # Final logs catch-up
    if [ -n "$log_file" ] && [ -f "$log_file" ]; then
        local final_logs=$(tail -c +$last_pos "$log_file")
        [ -n "$final_logs" ] && printf "\r\033[K%s" "$final_logs"
    fi

    # Final high-res time
    local end_ms=$(ruby -e 'puts (Time.now.to_f * 1000).to_i' 2>/dev/null || date +%s000)
    local final_elapsed=$((end_ms - start_ms))
    local final_sec=$((final_elapsed / 1000))
    local final_ms=$(((final_elapsed % 1000) / 100))
    
    # Store duration for caller
    G_DURATION="${final_sec}.${final_ms}s"
    
    # Show cursor
    printf "\033[?25h"
}
# Get current time in milliseconds
get_time_ms() {
    ruby -e 'puts (Time.now.to_f * 1000).to_i' 2>/dev/null || date +%s000
}
