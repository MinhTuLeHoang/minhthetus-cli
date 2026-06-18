# Git Tree

Visualizes the repository's commit graph.

If Atlassian SourceTree is installed on macOS, it will automatically open the current repository in SourceTree.

If SourceTree is not installed, or if the --cli flag is specified, it will display
a beautiful, colorized ASCII commit graph directly in your terminal using an interactive pager.

## Usage
```bash
minhthetus-cli git tree [options]
```

## Options

*   `-c, --cli`: Force displaying the colored Git commit tree in terminal (even if SourceTree is installed).
*   `-h, --help`: Show the help message and exit.

## Flow

1.  **SourceTree Check**:
    *   If `--cli` flag is not set, checks if Atlassian SourceTree is installed at `/Applications/SourceTree.app` or via `osascript`.
2.  **SourceTree Mode (macOS Default)**:
    *   Opens the current repository in SourceTree using `open -a SourceTree .`.
    *   Falls back to terminal tree display if SourceTree fails to open.
3.  **Terminal Graph Mode (Fallback/Forced)**:
    *   Renders a complete, colorized, and structured ASCII commit graph.
    *   Includes colored SHA hashes, relative commit times (e.g. `2 days ago`), commit messages, author names, and branch/tag decoration labels.
    *   Leverages the system pager (`less`) so you can interactively scroll, search, and navigate large git histories with full colors enabled.

## Version History

* **First Stable Version Supported**: `v1.2.0`
* **Latest Stable Version Update**: `v1.2.0`
