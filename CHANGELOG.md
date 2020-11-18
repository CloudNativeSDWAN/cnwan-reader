# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] (2020-09-24)

### Added

- This Changelog

- Readme: `Go Get` section

- Readme: `Clone The Project` section

- Readme: `Run as a Docker Container` section

- Readme: `Metadata Key` section

- Version command: `cnwan-reader version [--short|-s]`

### Changed

- Readme has been improved drastically with many sections being rewritten
in an effort to make it more understandable.

- `--metadata-key` is moved to `service directory`, as it would trigger an
error in `version`.

### Removed

- `COPYRIGHT` file, as all files created by the CN-WAN Reader `OWNERS` already
contain a copyright notice on top of them.

## [1.1.1] (2020-09-04)

### Fixed

- A concurrency issue preventing the program from receiving events while still
waiting for adaptor to replying is fixed.

## [1.1.0] (2020-09-02)

### Changed

- `--metadata-key` is now required

- `--credentials` is now moved under `servicedirectory` command and changed to
`--service-account` for better understing

- `--region` and `--project` are now marked as required in the framework and
thus not checked by the project anymore

- Readme: improve description on commands.

## [1.0.1] (2020-08-13)

### Fixed

- Trailing slashes are automatically removed from `--adaptor-api`

## [1.0.0] (2020-08-12)

### Added

- Core project
