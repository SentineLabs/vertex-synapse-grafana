# Changelog

All notable changes to the Vertex Synapse Grafana Datasource will be documented in this file.

## [1.0.0] - 2025-11-04

### Added
- Initial release of Vertex Synapse Grafana Datasource
- Support for executing Storm queries via `/api/v1/storm` (streaming) and `/api/v1/storm/call` (single result) endpoints
- Secure API key authentication for Cortex and Optic HTTP API endpoints
- Automatic Grafana time range injection as Storm variables
- Support for essential Storm data types: nodes, objects, lists, primitives

### Changed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Known Issues
- This is an initial release. Please report any issues on the GitHub repository.

[Unreleased]: https://github.com/sentinelabs/vertex-synapse-grafana/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/sentinelabs/vertex-synapse-grafana/releases/tag/v1.0.0
