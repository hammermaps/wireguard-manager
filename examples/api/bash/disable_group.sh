#!/bin/bash
# Disable all clients in a group

set -euo pipefail

# Check environment variables
if [ -z "${WIREGUARD_API_KEY:-}" ]; then
    echo "Error: WIREGUARD_API_KEY environment variable not set"
    exit 1
fi

if [ -z "${WIREGUARD_BASE_URL:-}" ]; then
    echo "Error: WIREGUARD_BASE_URL environment variable not set"
    exit 1
fi

GROUP="${1:-}"

if [ -z "$GROUP" ]; then
    echo "Usage: $0 <group_name>"
    echo "Example: $0 'Employees'"
    exit 1
fi

# Create JSON payload
payload=$(cat <<EOF
{
  "group": "$GROUP",
  "enabled": false
}
EOF
)

# Make API request
response=$(curl -s -X POST "${WIREGUARD_BASE_URL}/api/v1/group/set-status" \
    -H "Authorization: Bearer ${WIREGUARD_API_KEY}" \
    -H "Content-Type: application/json" \
    -d "$payload")

# Check response
if echo "$response" | grep -q '"success":true'; then
    echo "✓ Group disabled successfully!"
    if command -v jq &> /dev/null; then
        echo "$response" | jq -r '.message'
    else
        echo "$response"
    fi
else
    echo "✗ Error disabling group:"
    echo "$response"
    exit 1
fi
