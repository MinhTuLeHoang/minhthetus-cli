# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
