# Git Tree

The `git tree` command provides an interactive and highly readable visualization of your repository's commit graph. 

Depending on your environment, it will either open Atlassian SourceTree or render a beautiful ASCII graph in your terminal.

## Usage

```bash
minhthetus-cli git tree [flags]
```

### Flags
* `-c, --cli`: Forces the display of the colored commit graph in the terminal even if Atlassian SourceTree is installed.
* `-h, --help`: Displays help information for the command.

---

## Behavior

### 1. SourceTree Integration (macOS Default)
If Atlassian SourceTree is installed on your Mac, running `minhthetus-cli git tree`:
1. Checks for `/Applications/SourceTree.app` or its system registration.
2. Automatically opens the current repository in SourceTree using `open -a SourceTree .`.

### 2. High-Performance Terminal Graph (Fallback/Forced)
If SourceTree is not installed, or if the `--cli` flag is provided:
1. Renders a complete, colorized, and structured ASCII commit graph.
2. Leverages the system pager (`less`) so that you can interactively scroll, search, and navigate large git histories with full colors enabled.
3. Formats commit information with colored SHA hashes, relative commit times (e.g. `2 days ago`), commit messages, author names, and branch/tag decoration labels.

---

## Examples

* **Open current repository in SourceTree**:
  ```bash
  minhthetus-cli git tree
  ```

* **Force ASCII commit tree display in terminal**:
  ```bash
  minhthetus-cli git tree --cli
  ```

---

## Version History

* **First Stable Version Supported**: `v1.2.0`
* **Latest Stable Version Update**: `v1.2.0`
