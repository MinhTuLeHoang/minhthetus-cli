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
* **Latest Stable Version Update**: `v1.5.0`

- **v1.5.0**: Parallelized directory size calculations using a workload-balanced concurrency pool. Subdirectories at all depths are dynamically enqueued to ensure uniform worker utilization.
  - **Benchmark Data** (Tested on `/Users/lap15864-local` containing ~1.2 million directories and 150+ GB, warm filesystem cache):
    - **Single-Threaded Baseline**: `3m 12.292s` (192.29s)
    - **Parallel v1 (Static Top-Level split)**: `2m 8.197s` (128.20s) — `~1.50x` speedup
    - **Parallel v2 (Workload-Balanced)**: `1m 34.159s` (94.16s) — **`~2.04x` speedup** (more than 2x faster)
- **v1.3.4**: Introduced the directory size calculation command.
