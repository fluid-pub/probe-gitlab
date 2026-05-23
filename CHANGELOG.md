# Changelog — probe-gitlab

All notable changes to **fluid-pub/probe-gitlab** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Tag naming: `0.y.z` (no `v` prefix). Align `cmd/version.go` with the tag before release.

## [Unreleased]

### Changed

- Repository sync uses **go-git** instead of the `git` CLI (no `git` binary required in the distroless image).

## [0.1.0] - 2026-05-23

### Added

- **probe-core** integration: entity-based lifecycle (`users`, `groups`, `projects`, `repositories`, `code_files`), HTTP control plane (`/probes`), `runtime_config` merge, schema in image.
- CI/CD via `fluid-pub/actions` (test, release on semver tag, GHCR image `ghcr.io/fluid-pub/probe-gitlab`).
- Git repository sync and `code_files` indexing for RAG.

### Changed

- Configuration uses `state` (replacing `output`) and per-entity `refresh_interval` (replacing global probe refresh).
