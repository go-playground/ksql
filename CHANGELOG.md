# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.0] - 2023-05-31
### Added
- _string_ & _number_ COERCE types.

## [0.6.1] - 2023-05-25
### Fixed
- Fixed AND and OR not early exiting evaluation.

## [0.6.0] - 2023-01-05
### Added
- Added new `_uppercase_` & `_title_` COERCE identifiers.
- Added ability to use multiple COERCE identifiers at once separated by a comma.
- Added CLI ability to return original data if using an expression that returns a boolean.

## [0.5.1] - 2022-10-18
### Fixed
- Fixed CONTAINS_ANY for string contains comparisons with slice/array.

## [0.5.0] - 2022-10-13
### Added
- Added new `_lowercase_` COERCE identifier.

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

[Unreleased]: https://github.com/go-playground/ksql/compare/v0.7.0...HEAD
[0.6.1]: https://github.com/go-playground/ksql/compare/v0.7.0...v0.7.0
[0.6.1]: https://github.com/go-playground/ksql/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/go-playground/ksql/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/go-playground/ksql/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/go-playground/ksql/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/go-playground/ksql/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/go-playground/ksql/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/go-playground/ksql/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/go-playground/ksql/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/go-playground/ksql/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/go-playground/ksql/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/go-playground/ksql/commit/v0.1.0