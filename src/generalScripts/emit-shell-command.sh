#!/bin/bash
# Description: Utility to emit shell commands to the parent shell via integration pipe.

emit_shell_command() {
    if [ -n "$MINHTHETUS_SHELL_PIPE" ]; then
        echo "$1" >> "$MINHTHETUS_SHELL_PIPE"
    fi
}
