#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# The repository uses the v2 configuration schema. Build the binary with the
# selected Go toolchain so a cached/release binary built by Go 1.24 is never
# used for this Go 1.25 module.
lint_version="${GOLANGCI_LINT_VERSION:-v2.12.2}"
lint_module="github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
module_file="${GO_VERSION_FILE:-go.mod}"
required_go="${GO_MIN_VERSION:-}"

if [[ -z "${required_go}" ]] && [[ -f "${module_file}" ]]; then
  required_go="$(awk '/^go[[:space:]]/ { print $2; exit }' "${module_file}")"
fi
if [[ -z "${required_go}" ]]; then
  required_go="1.25.0"
fi

version_ge() {
  local current="$1"
  local required="$2"
  local lowest
  lowest="$(printf '%s\n%s\n' "${required}" "${current}" | sort -V | head -n 1)"
  [[ "${lowest}" == "${required}" ]]
}

go_version_from_bin() {
  local bin="$1"
  "$bin" version 2>/dev/null | awk '{print $3}' | sed 's/^go//'
}

go_bin_dir_from_bin() {
  local bin="$1"
  local gobin
  gobin="$(GOTOOLCHAIN=local "$bin" env GOBIN)"
  if [[ -n "${gobin}" ]]; then
    printf '%s\n' "${gobin}"
    return 0
  fi
  GOTOOLCHAIN=local "$bin" env GOPATH | awk '{print $1 "/bin"}'
}

check_go_binary() {
  local bin="$1"
  local ver
  if [[ ! -x "${bin}" ]]; then
    return 1
  fi
  ver="$(go_version_from_bin "${bin}")"
  if [[ -z "${ver}" ]]; then
    return 1
  fi
  if version_ge "${ver}" "${required_go}"; then
    GO_BIN_SELECTED="${bin}"
    GO_VERSION_SELECTED="${ver}"
    GO_BIN_DIR_SELECTED="$(go_bin_dir_from_bin "${bin}")"
    return 0
  fi
  return 1
}

ensure_and_select_go() {
  local bin

  GO_BIN_SELECTED=""
  GO_VERSION_SELECTED=""
  GO_BIN_DIR_SELECTED=""

  if [[ -n "${GO_BIN:-}" ]]; then
    check_go_binary "${GO_BIN}" && return 0
  fi

  bin="$(command -v go || true)"
  if [[ -n "${bin}" ]] && check_go_binary "${bin}"; then
    return 0
  fi

  if check_go_binary "/usr/local/go/bin/go"; then
    return 0
  fi

  if [[ -x "${script_dir}/ensure-go.sh" ]]; then
    GO_MIN_VERSION="${required_go}" GO_VERSION_FILE="${module_file}" "${script_dir}/ensure-go.sh" || true
    if check_go_binary "/usr/local/go/bin/go"; then
      return 0
    fi
  fi

  bin="$(command -v go || true)"
  if [[ -n "${bin}" ]] && check_go_binary "${bin}"; then
    return 0
  fi

  return 1
}

if ! ensure_and_select_go; then
  echo "::error::No suitable go toolchain for golangci-lint. required >= ${required_go}" >&2
  echo "::error::Check GO_BIN, or set GO_MIN_VERSION/GO_VERSION_FILE." >&2
  exit 1
fi

lint_bin="${GO_BIN_DIR_SELECTED}/golangci-lint"

lint_is_compatible() {
  local lint_output installed_lint installed_builder
  [[ -x "${lint_bin}" ]] || return 1

  lint_output="$("${lint_bin}" --version 2>&1)"
  installed_lint="v$(printf '%s\n' "${lint_output}" | sed -nE 's/.*version v?([0-9]+\.[0-9]+\.[0-9]+).*/\1/p' | head -n 1)"
  installed_builder="$(printf '%s\n' "${lint_output}" | sed -nE 's/.*built with (go[0-9]+\.[0-9]+\.[0-9]+).*/\1/p' | head -n 1)"

  [[ "${installed_lint}" == "${lint_version}" ]] || return 1
  [[ -n "${installed_builder}" ]] || return 1
  installed_builder="${installed_builder#go}"
  version_ge "${installed_builder}" "${required_go}"
}

if ! lint_is_compatible; then
  rm -f "${lint_bin}"
  GOTOOLCHAIN=local "${GO_BIN_SELECTED}" install "${lint_module}@${lint_version}"
fi

if ! lint_is_compatible; then
  echo "::error::golangci-lint must be ${lint_version} built with Go ${required_go}+" >&2
  exit 1
fi

"${lint_bin}" run ./...
