# StackRox MCP

## Project Overview

StackRox MCP is a Model Context Protocol (MCP) server that provides AI assistants with access to StackRox.

## Quick Start

Clone the repository:
```bash
git clone https://github.com/stackrox/stackrox-mcp.git
cd stackrox-mcp
```

Build the project:
```bash
go build -o stackrox-mcp ./cmd/stackrox-mcp
```

Run the server:
```bash
./stackrox-mcp
```

## How to Run Tests

Run all unit tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Generate detailed coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
