# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-11-25

### Added

- Initial release of Nashik Darshan SDK for Dart
- Support for all API endpoints:
  - Authentication (signup, login)
  - Places management (list, search, get by ID)
  - Categories management
  - Feed data (trending, popular, latest, nearby)
  - Reviews and ratings
  - User profile management
- Custom Dio instance support with interceptors
- Comprehensive documentation with examples
- Integration guide for Freezed, Clean Architecture, Bloc, and Cubits
- Type-safe models using built_value
- Full error handling support

### Technical Details

- Built with OpenAPI Generator
- Uses Dio for HTTP requests
- Uses built_value for immutable, type-safe models
- Compatible with Dart SDK >= 2.18.0
