#!/usr/bin/env python3
"""
Create a new WireGuard client
"""

import os
import sys
import requests
import argparse

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

def create_client(name: str, email: str, group: str, ip_range: str = "10.8.0.0/24"):
    """Create a new WireGuard client"""
    api_key, base_url = get_api_config()
    
    headers = {
        'Authorization': f'Bearer {api_key}',
        'Content-Type': 'application/json'
    }
    
    payload = {
        'name': name,
        'email': email,
        'group': group,
        'allocated_ips': [ip_range],
        'allowed_ips': ['0.0.0.0/0'],
        'use_server_dns': True,
        'enabled': True
    }
    
    try:
        response = requests.post(
            f'{base_url}/api/v1/client',
            headers=headers,
            json=payload,
            timeout=30
        )
        
        if response.status_code == 200:
            client = response.json()
            print(f"✓ Client '{name}' created successfully!")
            print(f"  ID: {client['id']}")
            print(f"  Public Key: {client['public_key']}")
            print(f"  Allocated IPs: {', '.join(client['allocated_ips'])}")
            return client
        else:
            error = response.json()
            print(f"✗ Error creating client: {error.get('message', 'Unknown error')}", file=sys.stderr)
            sys.exit(1)
            
    except requests.exceptions.RequestException as e:
        print(f"✗ Connection error: {e}", file=sys.stderr)
        sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description='Create a new WireGuard client')
    parser.add_argument('name', help='Client name')
    parser.add_argument('email', help='Client email')
    parser.add_argument('group', help='Client group')
    parser.add_argument('--ip', default='10.8.0.0/24', help='IP allocation (default: 10.8.0.0/24)')
    
    args = parser.parse_args()
    
    create_client(args.name, args.email, args.group, args.ip)

if __name__ == '__main__':
    main()
