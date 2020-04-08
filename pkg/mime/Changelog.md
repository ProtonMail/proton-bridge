# Do not modify this file!
It is here for historical reasons only. All changes should be documented in the
Changelog at the root of this repository.


# Changelog

## [2019-12-10] v1.0.2

### Added
* support for shift_JIS (cp932) encoding

## [2019-09-30] v1.0.1

### Changed
* fix divide by zero

## [2019-09-26] v1.0.0

### Changed
* Import-Export#192: filter header parameters
    * ignore twice the same parameter (take the latest)
    * convert non utf8 RFC2231 parameters to a single line utf8 RFC2231

