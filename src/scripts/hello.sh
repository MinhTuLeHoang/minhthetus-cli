#!/bin/bash
# Description: Simple greeting script that accepts a --name parameter.

# Simple parameter parsing for --name
NAME="User" # Default name

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --name)
      NAME="$2"
      shift
      ;;
  esac
  shift
done

echo "hello $NAME"
