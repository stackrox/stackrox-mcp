# Contributing to StackRox MCP

Thank you for your interest in contributing to StackRox MCP! This document provides guidelines and instructions for contributing to the project.

## Development Guidelines

### Code Quality Standards

All code must pass the following checks before being merged:

- **Formatting:** Run `go fmt ./...` to format your code
- **Linting:** Run `golint ./...` to check for style issues
- **Vetting:** Run `go vet ./...` to check for suspicious constructs
- **Testing:** All tests must pass with `go test ./...`

These checks are automatically run in CI for all pull requests.

### Testing

- Write unit tests for all new functionality
- Aim for 80% code coverage
- All error paths tested
- Run tests with coverage:
  ```bash
  go test -cover ./...
  ```
- Generate detailed coverage report:
  ```bash
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
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
