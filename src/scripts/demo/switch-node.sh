#!/bin/bash
# Description: Test switching Node version via shell integration pipe

VERSION=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --node)
      VERSION="$2"
      shift
      shift
      ;;
    *)
      shift
      ;;
  esac
done

if [ -z "$VERSION" ]; then
  echo "Error: Missing --node <version> argument." >&2
  exit 1
fi

if [ -n "$MINHTHETUS_SHELL_PIPE" ]; then
  echo "✦ Sending request to switch Node to $VERSION..." >&2
  
  # Write the nvm command to the integration pipe
  # The parent shell function will pick this up upon CLI exit
  echo "nvm use $VERSION" >> "$MINHTHETUS_SHELL_PIPE"
  
  echo "✅ Instruction sent. Node will switch once the CLI session ends." >&2
else
  echo "❌ Error: Shell integration pipe not detected." >&2
  echo "Please ensure you have run 'minhthetus-cli setup-completion' and sourced your shell config." >&2
  exit 1
fi
