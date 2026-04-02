#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"
DOC_FILE="${ROOT_DIR}/docs/model-evaluation.md"

# Validate required tools first
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    exit 1
fi

usage() {
    echo "Usage: $0 --model-id <id> --results <json-file>"
    echo ""
    echo "Update docs/model-evaluation.md with evaluation results from mcpchecker JSON output."
    echo ""
    echo "Options:"
    echo "  --model-id    Model identifier (e.g. gpt-5-mini)"
    echo "  --results     Path to mcpchecker JSON results file"
    echo "  -h, --help    Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --model-id gpt-5 --results e2e-tests/mcpchecker/mcpchecker-stackrox-mcp-e2e-out.json"
    exit 1
}

MODEL_ID=""
RESULTS_FILE=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --model-id)
            MODEL_ID="$2"
            shift 2
            ;;
        --results)
            RESULTS_FILE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Error: unknown option '$1'"
            usage
            ;;
    esac
done

if [[ -z "${MODEL_ID}" ]]; then
    echo "Error: --model-id is required"
    usage
fi

if [[ -z "${RESULTS_FILE}" ]]; then
    echo "Error: --results is required"
    usage
fi

if [[ ! -f "${RESULTS_FILE}" ]]; then
    echo "Error: results file not found: ${RESULTS_FILE}"
    exit 1
fi

if [[ ! -f "${DOC_FILE}" ]]; then
    echo "Error: documentation file not found: ${DOC_FILE}"
    exit 1
fi

TODAY=$(date +%Y-%m-%d)
START_MARKER="<!-- model:${MODEL_ID} start -->"
END_MARKER="<!-- model:${MODEL_ID} end -->"

# Generate the markdown block
generate_block() {
    local total passed
    total=$(jq 'length' "${RESULTS_FILE}")
    passed=$(jq '[.[] | select(.taskPassed == true)] | length' "${RESULTS_FILE}")
    local pct=$((100 * passed / total))

    echo "${START_MARKER}"
    echo ""
    echo "### ${MODEL_ID} — ${TODAY}"
    echo ""
    echo "**Overall: ${passed}/${total} tasks passed (${pct}%)**"
    echo ""
    echo "#### Task Results"
    echo ""
    echo "| # | Task | Result | toolsUsed | minCalls | maxCalls | Input Tokens | Output Tokens |"
    echo "|---|------|--------|-----------|----------|----------|--------------|---------------|"

    # Generate table rows
    jq -r '
        to_entries[] |
        .key as $i |
        .value |
        ($i + 1) as $num |
        .taskName as $name |
        (if .taskPassed then "Pass" else "**Fail**" end) as $result |
        (.assertionResults.toolsUsed // null) as $tu |
        (.assertionResults.minToolCalls // null) as $min |
        (.assertionResults.maxToolCalls // null) as $max |
        (if $tu == null then "\u2014"
         elif $tu.passed then "Pass"
         else "**Fail**"
         end) as $tuStr |
        (if $min == null then "\u2014"
         elif $min.passed then "Pass"
         else "**Fail**"
         end) as $minStr |
        (if $max == null then "\u2014"
         elif $max.passed then "Pass"
         else "**Fail**"
         end) as $maxStr |
        (.tokenEstimate.inputTokens) as $inputTokens |
        (.tokenEstimate.outputTokens) as $outputTokens |
        "| \($num) | \($name) | \($result) | \($tuStr) | \($minStr) | \($maxStr) | \($inputTokens) | \($outputTokens) |"
    ' "${RESULTS_FILE}"

    echo ""

    # Token totals
    local input_tokens output_tokens
    input_tokens=$(jq '[.[].tokenEstimate.inputTokens] | add' "${RESULTS_FILE}")
    output_tokens=$(jq '[.[].tokenEstimate.outputTokens] | add' "${RESULTS_FILE}")
    echo "**Total input tokens**: ${input_tokens} | **Total output tokens**: ${output_tokens}"
    echo ""
    echo "${END_MARKER}"
}

BLOCKFILE=$(mktemp)
TMPFILE=$(mktemp)
cleanup() { rm -f "${BLOCKFILE}" "${TMPFILE}"; }
trap cleanup EXIT

# shellcheck disable=SC2311
generate_block > "${BLOCKFILE}"

if grep -qF "${START_MARKER}" "${DOC_FILE}"; then
    # Update existing block: replace lines between markers (inclusive) with new block
    awk -v start="${START_MARKER}" -v end="${END_MARKER}" -v blockfile="${BLOCKFILE}" '
        $0 == start { skip=1; while ((getline line < blockfile) > 0) print line; next }
        $0 == end { skip=0; next }
        !skip { print }
    ' "${DOC_FILE}" > "${TMPFILE}"
    mv "${TMPFILE}" "${DOC_FILE}"

    echo "Updated existing results for ${MODEL_ID} in ${DOC_FILE}"
else
    # Insert new block before "## How to Run Evaluations"
    awk -v blockfile="${BLOCKFILE}" '
        /^## How to Run Evaluations/ {
            while ((getline line < blockfile) > 0) print line
            print ""
        }
        { print }
    ' "${DOC_FILE}" > "${TMPFILE}"
    mv "${TMPFILE}" "${DOC_FILE}"

    echo "Added new results for ${MODEL_ID} to ${DOC_FILE}"
fi
