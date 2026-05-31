#!/bin/bash
# Description: Automatically clones the GitHub Wiki, copies local wiki/ files, commits, and pushes.

set -e

# Configuration
WIKI_REPO="https://github.com/MinhTuLeHoang/minhthetus-cli.wiki.git"
TEMP_DIR=$(mktemp -d)

echo "⏳ Starting automated Wiki synchronization..."
echo "📂 Creating temporary workspace..."

# Clone the wiki repository
echo "⏳ Cloning Wiki repository..."
git clone "$WIKI_REPO" "$TEMP_DIR" --quiet

# Clean existing files in the cloned wiki repo except .git folder
echo "🧹 Cleaning old wiki workspace..."
find "$TEMP_DIR" -maxdepth 1 -not -name ".git" -not -name "$(basename "$TEMP_DIR")" -exec rm -rf {} +

# Copy local wiki files
echo "📂 Copying local wiki/ files to workspace..."
cp -R wiki/* "$TEMP_DIR/"

# Navigate to temp dir, stage, commit, and push
cd "$TEMP_DIR"

if [ -n "$(git status --porcelain)" ]; then
    echo "💾 Staging changes..."
    git add .
    
    echo "📝 Committing updates..."
    git commit -m "docs: sync wiki with latest codebase updates" --quiet
    
    echo "🚀 Pushing updates to GitHub..."
    git push origin master --quiet
    echo "✅ Wiki successfully synchronized!"
else
    echo "🕊 No documentation changes detected. Wiki is already up to date."
fi

# Cleanup
echo "🧹 Cleaning up temporary workspace..."
rm -rf "$TEMP_DIR"
