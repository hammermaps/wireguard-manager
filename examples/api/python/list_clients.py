#!/usr/bin/env python3
"""
List all WireGuard clients
"""

import os
import sys
import requests
from typing import List, Dict

def get_api_config() -> tuple:
    """Get API configuration from environment variables"""
    api_key = os.getenv('WIREGUARD_API_KEY')
    base_url = os.getenv('WIREGUARD_BASE_URL')
    
    if not api_key:
        print("Error: WIREGUARD_API_KEY environment variable not set", file=sys.stderr)
        sys.exit(1)
    
    if not base_url:
        print("Error: WIREGUARD_BASE_URL environment variable not set", file=sys.stderr)
        sys.exit(1)
    
    return api_key, base_url

def list_clients() -> List[Dict]:
    """Fetch all clients from the API"""
    api_key, base_url = get_api_config()
    
    headers = {
        'Authorization': f'Bearer {api_key}',
        'Content-Type': 'application/json'
    }
    
    try:
        response = requests.get(
            f'{base_url}/api/v1/clients',
            headers=headers,
            timeout=30
        )
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error fetching clients: {e}", file=sys.stderr)
        sys.exit(1)

def main():
    clients = list_clients()
    
    print(f"Total clients: {len(clients)}\n")
    
    # Group by status
    enabled_clients = [c for c in clients if c['Client']['enabled']]
    disabled_clients = [c for c in clients if not c['Client']['enabled']]
    
    print(f"Enabled: {len(enabled_clients)}")
    print(f"Disabled: {len(disabled_clients)}\n")
    
    # Group by group
    groups = {}
    for client_data in clients:
        client = client_data['Client']
        group = client['group'] or 'No Group'
        if group not in groups:
            groups[group] = []
        groups[group].append(client)
    
    # Print by group
    for group, group_clients in sorted(groups.items()):
        print(f"\n{group}: {len(group_clients)} clients")
        for client in group_clients:
            status = "✓" if client['enabled'] else "✗"
            ips = ', '.join(client['allocated_ips'])
            print(f"  {status} {client['name']} ({client['email']}) - {ips}")

if __name__ == '__main__':
    main()
