#!/bin/bash
# List all WireGuard clients

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

# Make API request
response=$(curl -s -X GET "${WIREGUARD_BASE_URL}/api/v1/clients" \
    -H "Authorization: Bearer ${WIREGUARD_API_KEY}" \
    -H "Content-Type: application/json")

# Check if jq is available for pretty printing
if command -v jq &> /dev/null; then
    echo "$response" | jq '.[] | {name: .Client.name, email: .Client.email, group: .Client.group, enabled: .Client.enabled, ips: .Client.allocated_ips}'
else
    echo "$response"
    echo ""
    echo "Tip: Install 'jq' for better output formatting"
fi
