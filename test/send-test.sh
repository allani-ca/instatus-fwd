#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
SECRET="${INSTATUS_WEBHOOK_SECRET:-}"

if [[ -z "$SECRET" ]]; then
    echo "Set INSTATUS_WEBHOOK_SECRET to the same value used by the service."
    exit 1
fi

PAYLOAD_FILE="$(dirname "$0")/payload-incident.json"

if [[ ! -f "$PAYLOAD_FILE" ]]; then
    echo "Payload file not found: $PAYLOAD_FILE"
    exit 1
fi

# The signature is HMAC-SHA256 over the exact bytes sent in the request.
# I think this is how the service expects the signature to be generated.
# https://superuser.com/revisions/1311623/2
SIGNATURE="$(
    openssl dgst -sha256 -hmac "$SECRET" "$PAYLOAD_FILE" |
    awk '{print $NF}'
)"

echo "POST $BASE_URL/webhook"
echo "Signature: $SIGNATURE"
echo

curl \
    --fail-with-body \
    -i \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-Instatus-Webhook-Signature: $SIGNATURE" \
    --data-binary "@$PAYLOAD_FILE" \
    "$BASE_URL/webhook"

echo
