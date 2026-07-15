# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.5.2] - 2026-07-15

### Added
- Show loading animation spinner immediately when executing size command
- Added recursive statistics counts summary (Files count, Directory count, and Total size) in a structured vertical bulleted list format

### Changed
- Display worker count in timing result print
- Dynamically colorize elapsed calculation duration based on speed thresholds
- Support non-TTY execution fallback to synchronous processing gracefully

## [1.5.1] - 2026-07-15

### Fixed
- Ensure shell RC file is created if it does not exist when setting up autocompletion

## [1.5.0] - 2026-07-15

### Added
- Workload-balanced parallel directory walking logic for speed optimization
- size_dev.go and size_prod.go build tag configuration

### Changed
- [DEV MODE] log in debug_register.go

### Fixed
- Resolved thread idle bottlenecks in directory traversal

## [1.4.2] - 2026-06-26

### Added
- Confirmation prompt when creating a new branch to choose the source branch (current or master)

## [1.4.1] - 2026-06-23

### Changed
- Updated the wiki sync GitHub Action workflow to use a custom 'WIKI_SYNC_TOKEN' instead of the reserved 'GITHUB_TOKEN'.
- Added support for manual workflow triggers ('workflow_dispatch') to the wiki sync GitHub Action workflow.

### Fixed
- Automatically run 'brew update' before upgrading the CLI via Homebrew to prevent using cached, outdated formulae databases.

## [1.4.0] - 2026-06-23

### Added
- Added sys clean-my-mac command to clean macOS caches and temp files

### Changed
- Colorized active Node version output in green

## [1.3.4] - 2026-06-20

### Added
- sys size command to list all files and folders in the current directory, sorted by size descending

### Changed
- updated CLI developer guide with custom logs styling recommendations

## [1.3.3] - 2026-06-15

### Added
- Added merge-publish release automation skill

### Changed
- Refactored publish helper scripts into subdirectories

### Fixed
- Corrected tag sorting logic in tag-dev-stg command

## [1.3.2] - 2026-06-03

### Added
- Pre-update version checking logic to verify if an update is needed before starting the download or compilation process.

## [1.3.1] - 2026-06-03

### Fixed
- Implement direct path switching for Node.js version in web commands to avoid incompatible node engine errors

## [1.3.0] - 2026-06-02

### Added
- Added a new built-in update command to automate CLI updates
- Added comprehensive Updating-the-CLI guide to the Wiki

### Changed
- Updated README.md with a new Keeping Up-to-Date section
- Updated publish skill instructions to enforce Makefile usage

## [1.2.0] - 2026-06-02

### Added
- Added git tree subcommand to visualize the repository's commit graph (opens SourceTree on macOS, or renders colorized ASCII graph in terminal with --cli)
- Added automatic developer documentation sync rules under .agents/rules/

### Changed
- Updated auto-complete options for the git tree subcommand

### Fixed
- Enforced AI compilation safety rules by migrating compiler safeguards

## [1.1.1] - 2026-06-01

### Changed
- Color-coded git branch names in git checkout command output
- Ran git checkout command in silent mode during discovery
- Documented make publish usage with ARGS variable

## [1.1.0] - 2026-06-01

### Added
- Dynamic autocomplete flag value suggestions for all subcommands (hello, checkout, merge-request, tag-dev-stg, switch-node)
- Instant flag-name autocomplete suggestions on empty <tab> input for all commands with flags

### Changed
- Guided Go install users to run setup-completion in README.md

## [1.0.3] - 2026-06-01

### Added
- Flat GitHub user wiki documentation under wiki/ directory
- Automated publish pipeline using GitHub CLI for Pull Request creation, merging, and tag pushing
- New make deploy-wiki synchronization utility
- New push-rc debug subcommand for release candidate bumping
- Centralized fallback version variables

### Removed
- All legacy shell scripts and old duplicate technical guides

## [1.0.2] - 2026-05-31

### Changed
- Re-architected Homebrew bump workflow to push directly to tap main branch.

## [1.0.1] - 2026-05-29

### Changed
- Synchronized completion installation and local folder structure configuration guides in documentation.

## [1.0.0] - 2026-05-29

### Added
- Complete native Go CLI migration utilizing Cobra command hierarchy and Bubble Tea UI helpers.
- Fully interactive `git checkout` integrating with JIRA ticket searching.
- Secure local Git account switching and metadata tracking.
- Zero-dependency deployment pipeline, completely replacing external `gum` and `jq` requirements.
