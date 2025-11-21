# Contributing to StackRox MCP

Thank you for your interest in contributing to StackRox MCP! This document provides guidelines and instructions for contributing to the project.

## Getting Started

Before contributing, get the project running locally:

### Initial Setup

Clone the repository:
```bash
git clone https://github.com/stackrox/stackrox-mcp.git
cd stackrox-mcp
```

Build the project:
```bash
make build
```

Run the server:
```bash
./stackrox-mcp
```

Once you have the project running, familiarize yourself with the development workflow below.

## Development Guidelines

### Code Quality Standards

All code must pass the following checks before being merged:

- **Formatting:** Run `make fmt` to format your code
- **Format Check:** Run `make fmt-check` to verify code is formatted
- **Linting:** Run `make lint` to check for style issues
- **Testing:** All tests must pass with `make test`

These checks are automatically run in CI for all pull requests.

### Available Make Targets

View all available make commands:
```bash
make help
```

Common development commands:
- `make build` - Build the binary
- `make test` - Run unit tests with coverage
- `make coverage-html` - Generate and view HTML coverage report
- `make fmt` - Format code
- `make fmt-check` - Check code formatting (fails if not formatted)
- `make lint` - Run golangci-lint
- `make clean` - Clean build artifacts and coverage files

### Testing

- Write unit tests for all new functionality
- Aim for 80% code coverage
- All error paths should be tested
- Run tests with coverage:
  ```bash
  make test
  ```
- Generate and view detailed coverage report:
  ```bash
  make coverage-html
  ```

## Pull Request Guidelines

### Creating a PR

- **Title:**
    - The title of your PR should be clear and descriptive.
    - It should be short enough to fit into the title box.
    - **PR addresses JIRA ticket:** `ROX-1234: Add feature ABC`
    - **Otherwise use conventional commit style:** `<type>(<scope>): <description>`
        - Types: `fix`, `docs`, `test`, `refactor`, `chore`, `ci`
        - Example: `fix(builds): Fix builds for ABC architecture`

- **Description:**
    - Describe the motivation for this change, or why some things were done a certain way.
    - Focus on what cannot be extracted from the code, e.g., alternatives considered and dismissed (and why), performance concerns, non-evident edge cases.

- **Validation:**
    - Provide information that can help the PR reviewer test changes and validate they are correct.
    - In some cases, it will be sufficient to mention that unit tests are added and they cover the changes.
    - In other cases, testing may be more complex, and providing steps on how to set up and test everything will be very valuable for reviewers.

### Merging a PR

- Make sure that **all CI statuses are green**.
- Always use `Squash and merge` as the merging mode (default).
- Double-check that the title of the commit ("subject line") is **your PR title**, followed by the PR number prefixed with a `#` in parentheses.
- Merge commit message example: `ROX-1234: Add feature ABC (#5678)`.
- The body of the commit message should be empty. If GitHub pre-populates it, delete it.

## Code Review Process

- All PRs require at least one approval before merging
- Address all reviewer comments and suggestions
- Keep PRs focused and reasonably sized
- Respond to feedback in a timely manner

## License

By contributing to StackRox MCP, you agree that your contributions will be licensed under the Apache License 2.0.
