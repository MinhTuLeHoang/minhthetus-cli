# Sys Host Command

The `sys host` command enables interactive search, view, addition, deletion, and formatting/sorting of the system host file (`/etc/hosts`).

## Usage

```bash
minhthetus-cli sys host
```

## Features

When you run the command, it opens an interactive menu using Bubble Tea with the following actions:

1. **Search Hosts**: Query `/etc/hosts` for domains or IPs. It displays exact matches along with relevant results (e.g. sharing the same IP or the same normalized domain).
2. **Show All Hosts**: Lists all parsed host entries cleanly, sorted according to the formatting rules.
3. **Add Host**: Interactively add a new IP-domain mapping. If you enter only one of the values, the CLI will prompt you for the missing piece. It automatically validates against duplicates.
4. **Delete Host**: Filters entries by query, prompts you to select an host, and requests confirmation before deleting it.
5. **Format & Sort hosts file**: Standardizes the layout of the `/etc/hosts` file based on the formatting rules.

## Formatting & Sorting Rules

When formatting or writing back changes, the file is standardized using these strict rules:

*   **One Domain Per Line**: Traditional multiple domains on a single line are split into distinct lines.
*   **Header Comments**: Contiguous comments or empty lines at the very top of `/etc/hosts` are preserved intact. All other inline/commented-out entries are discarded.
*   **IP Priorities**: Sorted by IP priority:
    1.  Loopback (`127.0.0.1`, `::1`, and other loopbacks).
    2.  Broadcast/Link-local (e.g. `255.255.255.255`, link-local unicast/multicast).
    3.  IPv4.
    4.  IPv6.
    5.  Invalid IP formats.
*   **Numerical Sorting**: For the same IP priority, sorted numerically by their IP byte values.
*   **Grouping & Domain Clustering**: Entries sharing the same IP address are grouped on consecutive lines (no blank lines in between) and sorted alphabetically by their normalized domain name (clustering environments of the same tool together, e.g. stripping environment prefixes like `stg-`, `qc-`, `dev-`, etc.).
*   **Spacing**: Different IP addresses are separated by a single blank line.

> [!IMPORTANT]
> Writing changes to `/etc/hosts` requires superuser (`sudo`) privileges. The tool writes the formatted content to a temporary file first and then executes `sudo cp` to overwrite the hosts file, prompting you for your macOS password if necessary.

## Version History

* **First Stable Version Supported**: `v1.5.3`
* **Latest Stable Version Update**: `v1.5.3`

- **v1.5.3**: Added interactive search, view, addition, deletion, and formatting/sorting of the `/etc/hosts` file.
