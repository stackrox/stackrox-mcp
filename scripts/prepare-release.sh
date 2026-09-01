#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
REPO_DIR="${SCRIPT_DIR}/.."
TEMPLATE_DIR="${SCRIPT_DIR}/templates"

VERSION=""
VERSION_DASHED=""
RELEASE_VERSION=""
RELEASE_VERSION_DASHED=""

usage() {
    cat <<EOF
Usage: $(basename "$0") <version>

Prepare a patch release for an existing Y-stream release branch.

Discovers the next Z version from existing tags, creates a prepare branch,
updates Chart.yaml, and generates Konflux Release resource YAMLs for
stage and production.

Arguments:
  version   Y-stream version in X.Y format (e.g., 0.2, 1.1)

Example:
  $(basename "$0") 0.2
  $(basename "$0") 1.1
EOF
    exit 0
}

log() {
    echo "==> $*"
}

error() {
    echo "ERROR: $*" >&2
    exit 1
}

parse_args() {
    local positional=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            -*)
                error "Unknown option: $1"
                ;;
            *)
                positional+=("$1")
                shift
                ;;
        esac
    done

    if [[ ${#positional[@]} -ne 1 ]]; then
        error "Expected 1 positional argument: <version>. Got ${#positional[@]}."
    fi

    export VERSION="${positional[0]}"
    export VERSION_DASHED="${VERSION//./-}"
}

validate_inputs() {
    if ! [[ "${VERSION}" =~ ^[0-9]+\.[0-9]+$ ]]; then
        error "Version must be in X.Y format (e.g., 0.2). Got: ${VERSION}"
    fi

    if ! command -v yq &>/dev/null; then
        error "yq is required but not found in PATH"
    fi

    if ! command -v git &>/dev/null; then
        error "git is required but not found in PATH"
    fi

    local yq_version
    yq_version=$(yq --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
    if ! printf '%s\n' "4.52.1" "${yq_version}" | sort -V | head -n1 | grep -q "4.52.1"; then
        error "yq version 4.52.1 or higher is required. Found: ${yq_version}"
    fi

    if [[ ! -d "${TEMPLATE_DIR}" ]]; then
        error "Templates directory not found: ${TEMPLATE_DIR}"
    fi

    git -C "${REPO_DIR}" fetch origin
    if ! git -C "${REPO_DIR}" rev-parse --verify "origin/release-${VERSION}" &>/dev/null; then
        error "Release branch origin/release-${VERSION} does not exist. Run start-y-stream-release.sh first."
    fi
}

checkout_release_branch() {
    local branch="release-${VERSION}"

    log "Checking out branch ${branch}..."
    git -C "${REPO_DIR}" checkout "${branch}"
    git -C "${REPO_DIR}" pull --ff-only origin "${branch}"
}

discover_next_version() {
    local latest_tag
    latest_tag=$(git -C "${REPO_DIR}" tag -l "${VERSION}.*" --sort=-version:refname | grep -E "^${VERSION}\.[0-9]+$" | head -1 || true)

    if [[ -z "${latest_tag}" ]]; then
        export RELEASE_VERSION="${VERSION}.0"
    else
        local patch
        patch="${latest_tag##*.}"
        export RELEASE_VERSION="${VERSION}.$((patch + 1))"
    fi

    export RELEASE_VERSION_DASHED="${RELEASE_VERSION//./-}"
    log "Next release version: ${RELEASE_VERSION}"
}

create_prepare_branch() {
    local branch="release-${RELEASE_VERSION_DASHED}-prepare"

    # -B resets the branch if it already exists, so reruns discard local unpushed commits.
    log "Creating branch ${branch}..."
    git -C "${REPO_DIR}" checkout -B "${branch}"
}

update_chart_version() {
    local chart="${REPO_DIR}/charts/stackrox-mcp/Chart.yaml"

    if [[ ! -f "${chart}" ]]; then
        error "Chart.yaml not found: ${chart}"
    fi

    log "Updating Chart.yaml to version ${RELEASE_VERSION}..."
    yq -i '.version = env(RELEASE_VERSION)' "${chart}"
    yq -i '.appVersion = env(RELEASE_VERSION)' "${chart}"
}

create_release_resources() {
    local releases_dir="${REPO_DIR}/releases/${RELEASE_VERSION}"

    log "Creating release resources in releases/${RELEASE_VERSION}/..."
    mkdir -p "${releases_dir}"

    export RELEASE_AUTHOR="${USER}"

    create_release_file "stage" "${releases_dir}"
    create_release_file "prod" "${releases_dir}"
}

create_release_file() {
    export RELEASE_ENV="$1"
    local output_dir="$2"
    local template="${TEMPLATE_DIR}/release-resource.yaml"
    local target="${output_dir}/agentic-cluster-security-suite-${VERSION_DASHED}--${RELEASE_VERSION_DASHED}--${RELEASE_ENV}.yaml"

    log "  Creating ${RELEASE_ENV} Release resource..."
    cp "${template}" "${target}"

    yq -i '.metadata.generateName |= (sub("X-Y-Z", env(RELEASE_VERSION_DASHED)) | sub("X-Y", env(VERSION_DASHED)) | sub("RELEASE_ENV", env(RELEASE_ENV)))' "${target}"
    yq -i '.metadata.labels."release.appstudio.openshift.io/author" = env(RELEASE_AUTHOR)' "${target}"
    yq -i '.spec.releasePlan |= (sub("RELEASE_ENV", env(RELEASE_ENV)) | sub("X-Y", env(VERSION_DASHED)))' "${target}"
    yq -i '(.spec.data.mapping.defaults.tags[] | select(test("X\\.Y\\.Z"))) |= sub("X\\.Y\\.Z", env(RELEASE_VERSION))' "${target}"
    yq -i '(.spec.data.mapping.defaults.tags[] | select(test("X\\.Y"))) |= sub("X\\.Y", env(VERSION))' "${target}"
    yq -i '(.spec.data.releaseNotes.references[] | select(test("X\\.Y\\.Z"))) |= sub("X\\.Y\\.Z", env(RELEASE_VERSION))' "${target}"
}

main() {
    parse_args "$@"
    validate_inputs

    log "Preparing release for Y-stream ${VERSION}"

    checkout_release_branch
    discover_next_version
    create_prepare_branch
    update_chart_version
    create_release_resources

    log ""
    log "Release preparation complete!"
    log "  Branch: release-${RELEASE_VERSION_DASHED}-prepare"
    log "  Chart.yaml: version=${RELEASE_VERSION}, appVersion=${RELEASE_VERSION}"
    log "  Stage:  releases/${RELEASE_VERSION}/agentic-cluster-security-suite-${VERSION_DASHED}--${RELEASE_VERSION_DASHED}--stage.yaml"
    log "  Prod:   releases/${RELEASE_VERSION}/agentic-cluster-security-suite-${VERSION_DASHED}--${RELEASE_VERSION_DASHED}--prod.yaml"
    log ""
    log "Review changes, then commit and push."
}

main "$@"
