# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [v2.0.0] - 2021-05-24

### Added

- storage: Add SSE part in formatFileObject (#17)
- storage: Add appender support (#18)
- storage: Implement CommitAppend (#19)
- *: Implement GSP-47 & GSP-51 (#21)
- storage: Implement GSP-61 Add object mode check for operations (#22)

### Changed

- docs: Update links to matrix
- storage: Idempotent storager delete operation (#20)
- *: Implement GSP-73 Organization rename (#23)

## [v1.1.0] - 2021-04-24

### Added

- storage: Add AccessTier, ContentType, ContentMD5 support for write (#4)
- *: Implement default pair support for service (#6)
- storage: Implement Create API (#10)
- *: Add UnimplementedStub (#11)
- tests: Introduce STORAGE_AZBLOB_INTEGRATION_TEST (#12)
- storage: Implement SSE support (#13)
- storage: Implement GSP-40 (#15)

### Changed

- ci: Only run Integration Test while push to master

## v1.0.0 - 2021-02-18

### Added

- Implement azblob services.

[v2.0.0]: https://github.com/beyondstorage/go-service-azblob/compare/v1.1.0...v2.0.0
[v1.1.0]: https://github.com/beyondstorage/go-service-azblob/compare/v1.0.0...v1.1.0