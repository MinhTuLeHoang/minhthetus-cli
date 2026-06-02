# minhthetus-cli Workspace Rules

This file defines workspace-scoped guidelines and rules for Google Antigravity AI agents operating in this repository.

## 🤖 COMPILER & SANDBOX SAFEGUARD
*   **DO NOT** run raw `go build`, `go run`, `go test`, or `go clean` commands directly.
*   Direct execution of `go` commands is treated as a sandboxed security constraint and will trigger a manual user approval prompt on every run.
*   **ALWAYS** use the pre-approved `make` targets instead:
    - **Build standard production binary**: `make build`
    - **Build developer debug binary**: `make build-dev`
    - **Build & install globally**: `make install`
    - **Run/Test locally**: `./minhthetus-cli <args>`
    - **Clean project**: `rm -f minhthetus-cli`
