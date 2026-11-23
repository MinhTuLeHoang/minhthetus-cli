# Sys Host Command

The `sys host` command enables interactive search, view, addition, deletion, and formatting/sorting of the system host file (`/etc/hosts` on macOS/Linux or `C:\Windows\System32\drivers\etc\hosts` on Windows).

## Usage

```bash
minhthetus-cli sys host
```

## Features

When you run the command, it opens an interactive menu using Bubble Tea with the following actions:

1. **Search Hosts**: Query hosts file for domains or IPs. It displays exact matches along with relevant results (e.g. sharing the same IP or the same normalized domain). When relevant results or matches are long, an interactive scrollable viewer opens starting at the top so exact matches are never pushed out of view.
2. **Show All Hosts**: Lists all parsed host entries cleanly, sorted according to the formatting rules (using the scroll viewer if list is long).
3. **Add Host**: Interactively add one or multiple IP-domain mappings. Supports single entries, multi-line pastes, and multiple host groups in a single input. If you enter only domains or only an IP, the CLI will prompt for the missing piece. Automatically validates against duplicates.
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
> Writing changes to the hosts file requires elevated privileges (`sudo` on macOS/Linux or Administrator on Windows). The tool writes the formatted content to a temporary file first and then overwrites the hosts file using elevated privileges (prompting for `sudo` on Unix or UAC prompt on Windows if not already elevated). After writing changes, the CLI automatically flushes the local DNS resolver cache (`ipconfig /flushdns` on Windows, `dscacheutil -flushcache` on macOS) so new mappings take effect immediately.

## Version History

* **First Stable Version Supported**: `v1.6.0`
* **Latest Stable Version Update**: `v1.12.1`

- **v1.12.1**: Added multi-host and multi-line input support in `Add Host`, unlimited input buffer length, and robust IP-domain stream grouping.
- **v1.11.4**: Added Windows support (`%SystemRoot%\System32\drivers\etc\hosts` with Administrator UAC elevation), automatic DNS cache flushing (`ipconfig /flushdns`), and interactive scrollable viewport for search results starting from the top.
- **v1.6.0**: Added interactive search, view, addition, deletion, and formatting/sorting of the `/etc/hosts` file.
