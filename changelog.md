# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Created `EnvironmentsHandler` to handle environment-specific requests.
- Added `apps/management/handlers/environments.go`.
- Added `apps/management/handlers/projects_test.go` for isolated project testing.

### Changed

- Separated project and environment management logic in `apps/management/handlers/projects.go`.
- Updated `apps/management/main.go` to route environment requests to the new `EnvironmentsHandler`.
- Updated `apps/management/handlers/environments_test.go` to use the new handler.
- Refactored `apps/management/handlers/flags_test.go` to support testing multiple handlers.

### Fixed

- Resolved a database insertion failure in environment tests caused by an invalid UUID in the `CreatedBy` field.
