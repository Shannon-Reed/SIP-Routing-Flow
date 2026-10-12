#!/usr/bin/env bash
# SIP Routing Flow - local demo (Go routing + Python analytics)
set -euo pipefail

GO_URL="${GO_URL:-http://localhost:8080}"
ANALYTICS_URL="${ANALYTICS_URL:-http://localhost:8000}"
ROUTE_REGION="${ROUTE_REGION:-US}"
FALLBACK_REGION="${FALLBACK_REGION:-APAC}"
CALLER="${CALLER:-+14045550123}"
DESTINATION="${DESTINATION:-+18005550100}"

for command in curl python3; do
  if ! command -v "$command" >/dev/null 2>&1; then
    printf 'Missing required command: %s\n' "$command" >&2
    exit 1
  fi
done

step() { printf '\n\033[1;36m%s\033[0m\n' "$1"; }
pretty() { python3 -m json.tool 2>/dev/null || cat; }

# Avoid exposing DB credentials or depending on a specific API response schema.
get() {
  local url="$1"
  curl --fail-with-body --silent --show-error --connect-timeout 5 --max-time 15 "$url"
}
post_route() {
  local region="$1"
  local call_id="$2"
  local payload

  payload=$(python3 -c '
import json, sys
print(json.dumps({
    "call_id": sys.argv[1],
    "caller": sys.argv[2],
    "destination": sys.argv[3],
    "region": sys.argv[4]
}))
' "$call_id" "$CALLER" "$DESTINATION" "$region")

  curl --silent --show-error \
    --connect-timeout 5 \
    --max-time 15 \
    --write-out '\nHTTP Status: %{http_code}\n' \
    -X POST "${GO_URL}/route" \
    -H "Content-Type: application/json" \
    -d "$payload"
}

step "1/5 - Checking both APIs"
printf 'Go API:      '
get "${GO_URL}/health" | pretty
printf 'Python API:  '
get "${ANALYTICS_URL}/health" | pretty

step "2/5 - Analytics summary BEFORE demo requests"
get "${ANALYTICS_URL}/analytics/summary" | pretty

# Unique IDs prevent confusion with previous demo runs.
run_id="$(date -u +%Y%m%dT%H%M%SZ)-$$"

step "3/5 - Region-specific routing (${ROUTE_REGION})"
printf 'Call ID: DEMO-REGION-%s\n' "$run_id"
post_route "$ROUTE_REGION" "DEMO-REGION-$run_id" | pretty

step "4/5 - Fallback routing (${FALLBACK_REGION})"
printf 'Call ID: DEMO-FALLBACK-%s\n' "$run_id"
post_route "$FALLBACK_REGION" "DEMO-FALLBACK-$run_id" | pretty

step "5/5 - Analytics AFTER demo requests"
printf '\nSummary:\n'
get "${ANALYTICS_URL}/analytics/summary" | pretty
printf '\nBy region:\n'
get "${ANALYTICS_URL}/analytics/regions" | pretty
printf '\nRecent decisions:\n'
get "${ANALYTICS_URL}/analytics/recent" | pretty

printf '\nDemo complete. Compare the BEFORE and AFTER results.\n'
printf 'Note: HTTP 503 means the requested region did not receive a route; inspect routing rules before claiming DEFAULT fallback works.\n'
