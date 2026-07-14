# Contributing to Sysgreet

Thank you for considering contributing to Sysgreet! This document outlines the development workflow and standards for this project.

## Code of Conduct

Be respectful, professional, and constructive. We welcome contributions from developers of all skill levels.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git
- Make (optional, for convenience commands)

### Getting Started

1. **Fork and clone the repository**

   ```bash
   git clone https://github.com/veteranbv/sysgreet.git
   cd sysgreet
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Run the project locally**

   ```bash
   go run ./cmd/sysgreet
   ```

4. **Run tests**

   ```bash
   make test
   # or
   go test ./...
   ```

5. **Run benchmarks**

   ```bash
   make bench
   # or
   go test -bench . ./test/benchmarks
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing code patterns and structure
   - Add tests for new functionality
   - Update documentation as needed

3. **Format and lint your code**

   ```bash
   make fmt
   make lint
   ```

4. **Run the full test suite**

   ```bash
   make test
   ```

5. **Commit your changes**

   ```bash
   git add .
   git commit -m "Brief description of changes"
   ```

   Commit message guidelines:
   - Use present tense ("Add feature" not "Added feature")
   - Keep the first line under 72 characters
   - Reference issues if applicable (#123)

### Submitting Changes

1. **Push to your fork**

   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a Pull Request**
   - Describe what the PR does and why
   - Reference any related issues
   - Ensure CI passes before requesting review

## Code Standards

### Principal Engineering Standards

This project follows strict engineering principles:

- **No shortcuts, stubs, or hardcoded values** - Build it right the first time
- **Clean, robust, and production-ready** - No halfway measures
- **Keep it tight** - Use the simplest solution that meets the need
- **Don't overengineer** - Every line of code should earn its place
- **When uncertain, ask** - Don't guess or make things up

### Go Code Guidelines

- **Follow Go idioms** - Use `gofmt`, `go vet`, and `golangci-lint`
- **Error handling** - Always handle errors; use graceful degradation where appropriate
- **Context usage** - Pass `context.Context` for cancellation support
- **Interfaces** - Design with interfaces for testability
- **No panics** - Handle errors gracefully; avoid `panic()` except in init functions
- **Documentation** - Add godoc comments for exported types and functions

### Testing Guidelines

- **Write tests for new features** - Aim for meaningful coverage
- **Table-driven tests** - Use subtests for multiple test cases
- **Integration tests** - Place in `test/integration/{platform}/` for OS-specific tests
- **Benchmarks** - Add benchmarks for performance-critical code
- **No flaky tests** - Tests must be deterministic and reliable

### Platform-Specific Code

When adding platform-specific code:

- Use build tags: `//go:build linux` or `//go:build windows`
- Place platform-specific implementations in separate files
- Maintain cross-platform parity where possible
- Test on all supported platforms (Linux, macOS, Windows)

## Project Structure

```text
sysgreet/
├── cmd/sysgreet/          # Main entry point
├── internal/
│   ├── ascii/             # ASCII art rendering
│   ├── banner/            # Banner orchestration and sections
│   ├── collectors/        # System data collectors
│   ├── config/            # Configuration loading
│   ├── network/           # Network utilities
│   ├── render/            # Layout, clipping, and JSON rendering
│   └── terminal/          # Terminal width/color detection, ANSI palette
├── test/
│   ├── benchmarks/        # Performance benchmarks
│   └── integration/       # Platform-specific integration tests
├── configs/               # Example configuration files
└── docs/                  # Documentation and examples
```

## Adding New Features

### Adding a New Collector

1. Define the interface in `internal/collectors/system.go`
2. Implement the collector in a new file or existing collector file
3. Add platform-specific implementations with build tags if needed
4. Register the collector in `cmd/sysgreet/main.go`
5. Add tests in `internal/collectors/*_test.go`

### Adding a New Banner Section

1. Create a builder in `internal/banner/sections_*.go`
2. Implement the `Builder` interface
3. Add the builder to `BuildersForConfig()` in `internal/banner/layout_config.go`
4. Update configuration structs if needed
5. Add tests for the new section

### Adding a New Configuration Option

1. Add the field to the config struct in `internal/config/defaults.go`
2. Update `mergeConfig()` in `internal/config/config.go`
3. Add environment variable support in `applyEnvOverrides()`
4. Update example configs in `configs/`
5. Document in README.md

## Performance Requirements

Sysgreet has strict performance requirements:

- **Startup time**: < 50ms median, < 80ms p95 (enforced by CI)
- **Binary size**: < 10MB for all platforms
- **Memory usage**: < 15MB RSS for default banner
- **No network activity**: All data must be collected locally

When adding new collectors or features, ensure they don't violate these constraints.

Collectors run concurrently inside `Providers.Gather` under a shared 250 ms
deadline, so honor the `context.Context` you receive — a collector that
ignores cancellation can still delay a login.

## Output Rules

- Never print wider than the terminal. The ascii renderer receives a
  `MaxWidth` and walks a fallback ladder; anything you add to the banner body
  is clipped by the render layer. Don't bypass either.
- All color must flow through `internal/terminal` so `NO_COLOR`, `TERM=dumb`,
  piped output, and Windows console quirks stay handled in one place.

## Documentation

### README Updates

Update the README if your changes:

- Add new configuration options
- Change behavior or output
- Add new installation methods
- Modify performance characteristics

### Code Documentation

- Add godoc comments to all exported functions, types, and constants
- Include examples in godoc when helpful
- Keep comments concise and accurate

## Release Process

Releases are handled by maintainers:

1. Update CHANGELOG.md with release notes
2. Create a version tag: `git tag v0.x.0`
3. Push the tag: `git push origin v0.x.0`
4. GitHub Actions automatically builds and publishes the release

## Getting Help

- **Questions?** Open a GitHub Discussion
- **Bug?** Open an issue with reproduction steps
- **Feature idea?** Open an issue describing the use case

## License

By contributing to Sysgreet, you agree that your contributions will be licensed under the Apache License 2.0.
