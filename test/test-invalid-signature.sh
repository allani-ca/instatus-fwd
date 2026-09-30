#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"

curl \
    -i \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-Instatus-Webhook-Signature: definitely-not-valid" \
    --data-binary @payload-incident.json \
    "$BASE_URL/webhook"
