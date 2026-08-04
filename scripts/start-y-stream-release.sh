#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
TEMPLATE_DIR="${SCRIPT_DIR}/templates"

VERSION=""
KONFLUX_DATA_REPO=""
VERSION_DASHED=""

usage() {
    cat <<EOF
Usage: $(basename "$0") <version> <konflux-data-repo>

Start a Y-stream release of agentic cluster security suite.

Creates release branches, updates Tekton pipelines, generates Konflux
configuration, and commits changes locally. Review the diff before pushing.

Arguments:
  version             Version in X.Y format (e.g., 1.5)
  konflux-data-repo   Path to the konflux-release-data repository

Options:
  -h, --help          Show this help message

Example:
  $(basename "$0") 1.5 /path/to/konflux-release-data
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

    if [[ ${#positional[@]} -ne 2 ]]; then
        error "Expected 2 positional arguments: <version> <konflux-data-repo>. Got ${#positional[@]}."
    fi

    export VERSION="${positional[0]}"
    KONFLUX_DATA_REPO="${positional[1]}"
    export VERSION_DASHED="${VERSION//./-}"
}

validate_inputs() {
    if ! [[ "${VERSION}" =~ ^[0-9]+\.[0-9]+$ ]]; then
        error "Version must be in X.Y format (e.g., 1.5). Got: ${VERSION}"
    fi

    if ! command -v yq &>/dev/null; then
        error "yq is required but not found in PATH"
    fi

    if ! command -v git &>/dev/null; then
        error "git is required but not found in PATH"
    fi

    if ! command -v kustomize &>/dev/null; then
        error "kustomize is required but not found in PATH (needed by build-single.sh)"
    fi

    if [[ ! -d "${KONFLUX_DATA_REPO}" ]]; then
        error "konflux-data-repo directory does not exist: ${KONFLUX_DATA_REPO}"
    fi

    if [[ ! -d "${KONFLUX_DATA_REPO}/.git" ]]; then
        error "konflux-data-repo is not a git repository: ${KONFLUX_DATA_REPO}"
    fi

    if [[ ! -d "${TEMPLATE_DIR}" ]]; then
        error "Templates directory not found: ${TEMPLATE_DIR}"
    fi

    local yq_version
    yq_version=$(yq --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
    # Get yq version, sort with 4.52.1 -> check that oldest version is 4.52.1.
    if ! printf '%s\n' "4.52.1" "${yq_version}" | sort -V | head -n1 | grep -q "4.52.1"; then
        error "yq version 4.52.1 or higher is required. Found: ${yq_version}"
    fi
}

prepare_stackrox_mcp_branch() {
    local branch="release-${VERSION}"

    log "Fetching origin in stackrox-mcp..."
    git -C "${SCRIPT_DIR}/.." fetch origin

    # -B resets the branch if it already exists, so reruns discard local unpushed commits.
    log "Creating branch ${branch} from origin/main..."
    git -C "${SCRIPT_DIR}/.." checkout -B "${branch}" origin/main
}

update_tekton_pipelines() {
    local tekton_dir="${SCRIPT_DIR}/../.tekton"
    local files=(
        "${tekton_dir}/acs-mcp-server-pull-request.yaml"
        "${tekton_dir}/acs-mcp-server-push-main.yaml"
        "${tekton_dir}/acs-mcp-server-push-release.yaml"
    )

    log "Updating Tekton pipeline files..."
    for file in "${files[@]}"; do
        if [[ ! -f "${file}" ]]; then
            error "Tekton file not found: ${file}"
        fi

        log "  Updating $(basename "${file}")..."

        yq -i --yaml-compact-seq-indent '.metadata.labels."appstudio.openshift.io/application" = "agentic-cluster-security-suite-" + env(VERSION_DASHED)' "${file}"
        yq -i --yaml-compact-seq-indent '.metadata.labels."appstudio.openshift.io/component" = "acs-mcp-server-" + env(VERSION_DASHED)' "${file}"
        yq -i --yaml-compact-seq-indent '.spec.taskRunTemplate.serviceAccountName = "build-pipeline-acs-mcp-server-" + env(VERSION_DASHED)' "${file}"
        yq -i --yaml-compact-seq-indent '(.spec.params[] | select(.name == "extra-labels") | .value[] | select(test("X\\.Y"))) |= sub("X\\.Y", env(VERSION))' "${file}"
    done
}

commit_stackrox_mcp() {
    local repo_dir="${SCRIPT_DIR}/.."

    git -C "${repo_dir}" add .tekton/
    if git -C "${repo_dir}" diff --cached --quiet; then
        log "No changes to commit in stackrox-mcp (already up to date)"
    else
        log "Committing Tekton pipeline changes..."
        git -C "${repo_dir}" commit -m "Configure Konflux pipelines for release ${VERSION}"
    fi
}

prepare_konflux_data_branch() {
    local branch="agentic-suite-release-${VERSION_DASHED}"

    log "Fetching origin in konflux-release-data..."
    git -C "${KONFLUX_DATA_REPO}" fetch origin

    # -B resets the branch if it already exists, so reruns discard local unpushed commits.
    log "Creating branch ${branch} from origin/main..."
    git -C "${KONFLUX_DATA_REPO}" checkout -B "${branch}" origin/main
}

create_release_plan_admissions() {
    local rpa_dir="${KONFLUX_DATA_REPO}/config/kflux-prd-rh02.0fk9.p1/product/ReleasePlanAdmission/agentic-cluster-security-suite"

    if [[ ! -d "${rpa_dir}" ]]; then
        error "ReleasePlanAdmission directory not found: ${rpa_dir}"
    fi

    create_rpa_file "stage"
    create_rpa_file "prod"
}

create_rpa_file() {
    export RPA_ENV="$1"
    local rpa_dir="${KONFLUX_DATA_REPO}/config/kflux-prd-rh02.0fk9.p1/product/ReleasePlanAdmission/agentic-cluster-security-suite"
    local template="${TEMPLATE_DIR}/release-plan-admission-${RPA_ENV}.yaml"
    local target="${rpa_dir}/agentic-cluster-security-suite-${RPA_ENV}-${VERSION_DASHED}.yaml"

    log "Creating ${RPA_ENV} ReleasePlanAdmission..."
    cp "${template}" "${target}"

    yq -i '.metadata.name = "agentic-cluster-security-suite-" + env(RPA_ENV) + "-" + env(VERSION_DASHED)' "${target}"
    yq -i '.spec.applications[0] = "agentic-cluster-security-suite-" + env(VERSION_DASHED)' "${target}"
    yq -i '.spec.data.releaseNotes.product_version = env(VERSION)' "${target}"
    yq -i '.spec.data.mapping.components[0].name = "acs-mcp-server-" + env(VERSION_DASHED)' "${target}"
    yq -i '(.spec.data.mapping.defaults.tags[] | select(test("X\\.Y"))) |= sub("X\\.Y", env(VERSION))' "${target}"
}

create_release_kustomization() {
    local tenant_base="${KONFLUX_DATA_REPO}/tenants-config/cluster/kflux-prd-rh02/tenants/agentic-cluster-security-suite-tenant/agentic-cluster-security-suite"
    local release_dir="${tenant_base}/release-${VERSION}"

    log "Creating release-${VERSION} kustomization directory..."
    mkdir -p "${release_dir}"

    cp "${TEMPLATE_DIR}/release-kustomization.yaml" "${release_dir}/kustomization.yaml"

    local target="${release_dir}/kustomization.yaml"
    yq -i '.patches[0].patch |= sub("release-X\\.Y", "release-" + env(VERSION))' "${target}"
    yq -i '.transformers[0] |= sub("-X-Y", "-" + env(VERSION_DASHED))' "${target}"
}

update_parent_kustomization() {
    local parent_kustomization="${KONFLUX_DATA_REPO}/tenants-config/cluster/kflux-prd-rh02/tenants/agentic-cluster-security-suite-tenant/agentic-cluster-security-suite/kustomization.yaml"

    if [[ ! -f "${parent_kustomization}" ]]; then
        error "Parent kustomization.yaml not found: ${parent_kustomization}"
    fi

    log "Adding release-${VERSION} to parent kustomization.yaml..."
    yq -i '.resources |= (. + ["release-" + env(VERSION)] | unique)' "${parent_kustomization}"
}

run_kustomize_build() {
    local tenants_config_dir="${KONFLUX_DATA_REPO}/tenants-config"

    if [[ ! -f "${tenants_config_dir}/build-single.sh" ]]; then
        error "build-single.sh not found in: ${tenants_config_dir}"
    fi

    log "Running kustomize build validation..."
    (cd "${tenants_config_dir}" && bash ./build-single.sh agentic-cluster-security-suite-tenant)
}

commit_konflux_data() {
    git -C "${KONFLUX_DATA_REPO}" add \
        config/kflux-prd-rh02.0fk9.p1/product/ReleasePlanAdmission/agentic-cluster-security-suite/ \
        tenants-config/cluster/kflux-prd-rh02/tenants/agentic-cluster-security-suite-tenant/ \
        tenants-config/auto-generated/cluster/kflux-prd-rh02/tenants/agentic-cluster-security-suite-tenant
    if git -C "${KONFLUX_DATA_REPO}" diff --cached --quiet; then
        log "No changes to commit in konflux-release-data (already up to date)"
    else
        log "Committing changes in konflux-release-data..."
        git -C "${KONFLUX_DATA_REPO}" commit -m "Add agentic-cluster-security-suite release ${VERSION} configuration"
    fi
}

main() {
    parse_args "$@"
    validate_inputs

    log "Starting Y-stream release ${VERSION} (dashed: ${VERSION_DASHED})"

    log ""
    log "=== Part 1: stackrox-mcp ==="
    prepare_stackrox_mcp_branch
    update_tekton_pipelines
    commit_stackrox_mcp

    log ""
    log "=== Part 2: konflux-release-data ==="
    prepare_konflux_data_branch
    create_release_plan_admissions
    create_release_kustomization
    update_parent_kustomization
    run_kustomize_build
    commit_konflux_data

    log ""
    log "Y-stream release ${VERSION} setup complete!"
    log "Review changes before pushing:"
    log "  git -C ${SCRIPT_DIR}/.. diff HEAD~1"
    log "  git -C ${KONFLUX_DATA_REPO} diff HEAD~1"
}

main "$@"
