# Third-Party Notices

This repository includes data and compatibility behavior derived from third-party projects.

## Matomo Device Detector

- Component: Matomo Device Detector regex and taxonomy data
- Upstream: https://github.com/matomo-org/device-detector
- License: LGPL-3.0-or-later
- Local usage:
  - Pinned source snapshot inputs in `sync/snapshots/v1/`
  - Deterministic generated artifact in `sync/compiled.json`
  - Provenance metadata in `compliance/provenance.json`
- License copy: `licenses/LGPL-3.0-or-later.txt`

## Redistribution Notes

- Preserve this notice file and the corresponding license files when redistributing binaries or source.
- Keep provenance metadata (`compliance/provenance.json`) with release artifacts so generated data remains auditable.
