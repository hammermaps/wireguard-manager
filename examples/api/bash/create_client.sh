#!/bin/bash
# Create a new WireGuard client

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

# Get parameters
NAME="${1:-}"
EMAIL="${2:-}"
GROUP="${3:-}"
IP="${4:-10.8.0.0/24}"

if [ -z "$NAME" ] || [ -z "$EMAIL" ] || [ -z "$GROUP" ]; then
    echo "Usage: $0 <name> <email> <group> [ip_range]"
    echo "Example: $0 'John Doe' 'john@example.com' 'Employees' '10.8.0.50/32'"
    exit 1
fi

# Create JSON payload
payload=$(cat <<EOF
{
  "name": "$NAME",
  "email": "$EMAIL",
  "group": "$GROUP",
  "allocated_ips": ["$IP"],
  "allowed_ips": ["0.0.0.0/0"],
  "use_server_dns": true,
  "enabled": true
}
EOF
)

# Make API request
response=$(curl -s -X POST "${WIREGUARD_BASE_URL}/api/v1/client" \
    -H "Authorization: Bearer ${WIREGUARD_API_KEY}" \
    -H "Content-Type: application/json" \
    -d "$payload")

# Check response
if echo "$response" | grep -q '"id"'; then
    echo "✓ Client created successfully!"
    if command -v jq &> /dev/null; then
        echo "$response" | jq '{id, name, email, group, public_key, allocated_ips}'
    else
        echo "$response"
    fi
else
    echo "✗ Error creating client:"
    echo "$response"
    exit 1
fi
