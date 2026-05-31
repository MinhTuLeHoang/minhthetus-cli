# Utils QR Generator (Remote)

Generates a QR code from text or a URL directly in your terminal. This is a remote command designed to keep the main CLI package lightweight.

## Usage
```bash
minhthetus-cli utils genQR <content> [options]
```

## Options

| Option | Description |
| :--- | :--- |
| `-m`, `--mode <type>` | Generation mode: `api` (default) or `self-implement`. |

## Flow

1.  **Parse Arguments**:
    *   Reads the content to be encoded.
    *   Determines if the external API or local engine should be used via the `--mode` flag.
2.  **Mode Resolution**:
    *   **API Mode (Default)**:
        *   Constructs a request to `https://qrenco.de/`.
        *   Displays the raw ANSI output received from the service.
    *   **Self-Implement**:
        *   Encodes the data into a QR matrix using a local implementation (no external dependencies).
        *   Renders the matrix using "Half-Block" characters for a perfect 1:1 square aspect ratio.
3.  **Display Output**:
    *   Prints the QR code to the standard output.
    *   Provides a success message upon completion.

## Version History
* **First Stable Version Supported**: `v1.0.0`
* **Latest Stable Version Update**: `v1.0.0`
