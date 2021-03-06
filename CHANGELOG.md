# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2022-07-29
### Added
- The ability for CONTAINS_ANY and CONTAINS_ALL to check if a String contains any|all of the values
  within an Array. Any non-string data types return a false.

## [0.3.2] - 2022-07-29
### Fixed
- && and || expression chaining.

## [0.3.1] - 2022-07-19
### Fixed
- Fixed number parsing for exponential numbers eg. 1e10.

## [0.3.0] - 2022-07-19
### Added
- Added BETWEEN operator support <value> BETWEEN <value> <value>

### Fixed
- Missing Gt, Gte, Lt, Lte for DateTime data type.

## [0.2.0] - 2022-07-18
### Fixed
- Reworked Parsing algorithm fixing a bunch of scoping issues.
- Added COERCE to DateTime support.
- Added CONTAINS_ANY & CONTAINS_ALL operators.

## [0.1.2] - 2022-06-08
### Fixed
- Handling of commas in arrays.

## [0.1.1] - 2022-06-08
### Fixed
- CLI output to be JSON.

## [0.1.0] - 2022-03-23
### Added
- Initial conversion from https://github.com/rust-playground/ksql.

[Unreleased]: https://github.com/go-playground/ksql/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/go-playground/ksql/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/go-playground/ksql/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/go-playground/ksql/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/go-playground/ksql/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/go-playground/ksql/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/go-playground/ksql/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/go-playground/ksql/commit/v0.1.0