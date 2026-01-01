# WireGuard Manager API Examples

This directory contains ready-to-use example scripts for interacting with the WireGuard Manager API.

## Setup

Before using these examples, you need to:

1. Create an API key in the WireGuard Manager web interface (API → API Key Management)
2. Set the required environment variables:

```bash
export WIREGUARD_API_KEY="your_api_key_here"
export WIREGUARD_BASE_URL="https://your-server.com"
```

## Available Examples

### Bash Scripts

- **[list_clients.sh](bash/list_clients.sh)** - List all VPN clients
- **[create_client.sh](bash/create_client.sh)** - Create a new VPN client
- **[disable_group.sh](bash/disable_group.sh)** - Disable all clients in a group
- **[batch_create_clients.sh](bash/batch_create_clients.sh)** - Create multiple clients from a CSV file

### Python Scripts

- **[list_clients.py](python/list_clients.py)** - List and filter clients
- **[create_client.py](python/create_client.py)** - Create a client with error handling
- **[wireguard_api_client.py](python/wireguard_api_client.py)** - Complete API client library
- **[disable_inactive_clients.py](python/disable_inactive_clients.py)** - Disable clients inactive for X days

### JavaScript/Node.js Scripts

- **[list_clients.js](nodejs/list_clients.js)** - List all clients
- **[create_client.js](nodejs/create_client.js)** - Create a new client
- **[wireguard_api_client.js](nodejs/wireguard_api_client.js)** - Reusable API client module

### Go Scripts

- **[wireguard_client.go](go/wireguard_client.go)** - Complete Go API client
- **[list_clients.go](go/list_clients.go)** - Simple client listing example

### PowerShell Scripts

- **[WireGuardAPI.psm1](powershell/WireGuardAPI.psm1)** - PowerShell module for API access
- **[List-Clients.ps1](powershell/List-Clients.ps1)** - List all clients
- **[New-Client.ps1](powershell/New-Client.ps1)** - Create a new client

## Usage

### Bash Example

```bash
cd examples/api/bash
export WIREGUARD_API_KEY="your_key"
export WIREGUARD_BASE_URL="https://your-server.com"
./list_clients.sh
```

### Python Example

```bash
cd examples/api/python
pip install requests
export WIREGUARD_API_KEY="your_key"
export WIREGUARD_BASE_URL="https://your-server.com"
python list_clients.py
```

### Node.js Example

```bash
cd examples/api/nodejs
npm install axios
export WIREGUARD_API_KEY="your_key"
export WIREGUARD_BASE_URL="https://your-server.com"
node list_clients.js
```

### Go Example

```bash
cd examples/api/go
export WIREGUARD_API_KEY="your_key"
export WIREGUARD_BASE_URL="https://your-server.com"
go run list_clients.go
```

### PowerShell Example

```powershell
cd examples/api/powershell
$env:WIREGUARD_API_KEY = "your_key"
$env:WIREGUARD_BASE_URL = "https://your-server.com"
Import-Module .\WireGuardAPI.psm1
Get-WGClients
```

## CSV Format for Batch Operations

For batch client creation, use this CSV format:

```csv
name,email,group
John Doe,john@example.com,Employees
Jane Smith,jane@example.com,Employees
Bob Johnson,bob@example.com,Contractors
```

## Security Note

Never hardcode API keys in your scripts. Always use environment variables or secure secret management systems.

## Documentation

For complete API documentation, see:
- [API_DOCUMENTATION.md](../../API_DOCUMENTATION.md) (English)
- [API_DOCUMENTATION_DE.md](../../API_DOCUMENTATION_DE.md) (Deutsch)
