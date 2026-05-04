#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8080}"
USERNAME="${2:-admin}"
PASSWORD="${3:-admin123}"

payload=$(cat <<JSON
{"username":"${USERNAME}","password":"${PASSWORD}"}
JSON
)

resp=$(curl -sS -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "${payload}")

echo "$resp"

if echo "$resp" | rg -q '"code"\s*:\s*0' && echo "$resp" | rg -q '"token"\s*:'; then
  echo "login smoke: PASS"
  exit 0
fi

echo "login smoke: FAIL"
exit 1
