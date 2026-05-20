#!/usr/bin/env bash
# Loads SEVENTHINGS_* credentials from .env and runs the integration test suite.
#
# Usage:
#   scripts/run-integration.sh                    # run all integration tests
#   scripts/run-integration.sh -run TestPerson    # forward args to `go test`
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$REPO_ROOT/.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "error: $ENV_FILE not found. Copy .env.example to .env and fill in values." >&2
  exit 1
fi

# Export each KEY=VALUE line from .env (skips comments and blank lines).
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

missing=()
for var in SEVENTHINGS_BASE_URL SEVENTHINGS_USERNAME SEVENTHINGS_PASSWORD SEVENTHINGS_CLIENT_ID; do
  if [[ -z "${!var:-}" ]]; then
    missing+=("$var")
  fi
done
if [[ ${#missing[@]} -gt 0 ]]; then
  echo "error: missing required variables in .env: ${missing[*]}" >&2
  exit 1
fi

cd "$REPO_ROOT"
exec go test -tags=integration -v "$@" ./...
