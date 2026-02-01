# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of dromos-authkit
- JWT authentication middleware for Gin applications
- JWKS caching with automatic refresh
- Role-based access control (RBAC) middleware
- Multi-tenant support with organization context
- WebSocket token extraction support
- Helper functions for accessing user claims
- Configurable path skipping for public routes

### Security

- RSA signature verification for JWTs
- Audience validation to prevent token reuse
- Automatic token expiration checking

## [1.0.0] - YYYY-MM-DD

### Added

- First stable release

---

## Release Process

To create a new release:

1. Update this CHANGELOG.md with the new version and release date
2. Commit the changes: `git commit -am "chore: prepare release vX.Y.Z"`
3. Push to main: `git push origin main`
4. Create a tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
5. Push the tag: `git push origin vX.Y.Z`
6. GitHub Actions will automatically create the release

Or use the GitHub Actions workflow:

- Go to Actions → Tag → Run workflow
- Enter the version (e.g., v1.0.0)
