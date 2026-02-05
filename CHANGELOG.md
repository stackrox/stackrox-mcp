# Changelog

All notable changes to this project will be documented in this file.

## [Next release]

## [0.1.0]

### Added
- Initial developer preview release
- MCP server implementation with HTTP and stdio transports
- Vulnerability management tools:
  - `get_clusters_with_orchestrator_cve` - Find CVEs in Kubernetes components
  - `get_deployments_for_cve` - Find CVEs in application workloads
  - `get_nodes_for_cve` - Find CVEs in node OS packages
  - `list_clusters` - List all managed clusters
- Configuration via YAML files and environment variables
- Authentication support: passthrough and static modes
- TLS support with certificate verification
- HTTP/1 fallback mode for restricted network environments
- Retry logic with exponential backoff
- Multi-architecture container images (amd64, arm64, ppc64le, s390x)
- Kubernetes/OpenShift deployment via Helm chart
- OpenShift Route support with TLS

### Documentation
- Comprehensive README with quick start, configuration, and deployment guides
- Architecture documentation
- OpenShift Lightspeed integration guide
- Configuration examples
- Testing and verification guides
- Helm chart deployment instructions

[0.1.0]: https://github.com/stackrox/stackrox-mcp/releases/tag/v0.1.0
