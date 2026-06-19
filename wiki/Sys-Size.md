# System Size Command

Lists the size of all files and folders in the current directory, sorted by size descending.

## Usage
```bash
minhthetus-cli sys size [options]
```

## Options

*   `-h, --help`: Show the help message and exit.

## Flow

1.  **Retrieve Current Working Directory**:
    *   Finds the absolute path of the current directory.
2.  **Scan Directory Entries**:
    *   Reads all files and folders, including hidden files/directories (e.g. starting with `.`).
3.  **Calculate Disk Usage**:
    *   For files: Retrieves the file size directly using standard file metadata.
    *   For directories: Recursively walks the directory tree to calculate the total size of all nested files.
4.  **Sort Entries**:
    *   Sorts all directory entries descending by their calculated size.
5.  **Render Table**:
    *   Presents a clean, formatted table mapping: `SIZE`, `TYPE` (indicating `dir` or `file`), and `NAME`.
    *   Colorizes size outputs according to disk space:
        *   **Red**: Size $\ge$ 100 MB.
        *   **Yellow**: Size $\ge$ 1 MB.
        *   **Green**: Size $<$ 1 MB.
    *   Displays the total time taken to calculate the sizes (e.g. `Calculated in 2.3ms`).

## Version History

* **First Stable Version Supported**: `v1.3.4`
* **Latest Stable Version Update**: `v1.3.4`
