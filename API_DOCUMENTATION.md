# WireGuard Manager API Documentation

This document describes the REST API features added to WireGuard Manager, including API key management and group operations.

## Features

1. **API Key Management**: Create and manage API keys for external access
2. **Group Operations**: Bulk enable/disable all clients in a group
3. **API Statistics**: Track API usage with detailed logs
4. **Permission-based Access**: Control what operations each API key can perform

## Getting Started

### 1. Creating an API Key

Navigate to **API → API Key Management** in the web interface.

1. Click "New API Key"
2. Enter a name for the key
3. Select permissions:
   - `read:clients` - View client information
   - `write:clients` - Create, update, and delete clients
   - `read:server` - View server configuration
   - `write:server` - Modify server configuration
   - `manage:groups` - Enable/disable groups of clients
   - `read:stats` - View API statistics
4. Click "Create"
5. **Important**: Copy the API key immediately - it will only be shown once!

### 2. Using the API

All external API endpoints are under `/api/v1/` and require authentication via Bearer token.

#### Authentication

Include the API key in the Authorization header:

```bash
Authorization: Bearer YOUR_API_KEY
```

## API Endpoints

### Client Operations

#### List All Clients
```bash
GET /api/v1/clients
Authorization: Bearer YOUR_API_KEY
```

**Required Permission**: `read:clients`

**Response**:
```json
[
  {
    "Client": {
      "id": "...",
      "name": "Client 1",
      "email": "client1@example.com",
      "group": "GroupA",
      "enabled": true,
      "allocated_ips": ["10.8.0.2/32"],
      "allowed_ips": ["0.0.0.0/0"]
    },
    "QRCode": ""
  }
]
```

#### Get Single Client
```bash
GET /api/v1/client/:id
Authorization: Bearer YOUR_API_KEY
```

**Required Permission**: `read:clients`

#### Create Client
```bash
POST /api/v1/client
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "name": "New Client",
  "email": "newclient@example.com",
  "group": "GroupA",
  "allocated_ips": ["10.8.0.10/32"],
  "allowed_ips": ["0.0.0.0/0"],
  "use_server_dns": true,
  "enabled": true
}
```

**Required Permission**: `write:clients`

#### Update Client
```bash
PUT /api/v1/client
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "id": "client_id_here",
  "name": "Updated Name",
  "email": "updated@example.com",
  "group": "GroupB",
  "enabled": true
  // ... other fields
}
```

**Required Permission**: `write:clients`

#### Delete Client
```bash
DELETE /api/v1/client/:id
Authorization: Bearer YOUR_API_KEY
```

**Required Permission**: `write:clients`

#### Set Client Status
```bash
POST /api/v1/client/set-status
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "id": "client_id_here",
  "status": false
}
```

**Required Permission**: `write:clients`

### Group Operations

#### Enable/Disable All Clients in a Group
```bash
POST /api/v1/group/set-status
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "group": "GroupName",
  "enabled": false
}
```

**Required Permission**: `manage:groups`

**Response**:
```json
{
  "success": true,
  "message": "Successfully disabled 5 client(s) in group 'GroupName'"
}
```

This is useful for:
- Temporarily disabling access for a team or department
- Emergency shutdown of a group of clients
- Maintenance windows

## Web Interface Features

### Group Management

In the main clients view, clients are automatically grouped. Each group has control buttons:

- **Enable All**: Enable all clients in the group
- **Disable All**: Disable all clients in the group

### API Key Management

Access via **API → API Key Management**:

- View all API keys with their permissions
- Create new API keys
- Enable/disable API keys
- Update API key permissions
- Delete API keys
- See last used timestamp

### API Statistics

Access via **API → API Statistics**:

- Total API keys count
- Active API keys count
- Total API calls
- Calls today
- Recent access logs with:
  - Timestamp
  - API key name
  - HTTP method
  - Endpoint
  - IP address
  - User agent
  - Response status
- Usage chart by API key

## Security Best Practices

1. **Store API Keys Securely**: Never commit API keys to version control
2. **Use Minimal Permissions**: Only grant the permissions required for each use case
3. **Rotate Keys Regularly**: Create new keys and delete old ones periodically
4. **Monitor Usage**: Check API statistics regularly for unusual activity
5. **Disable Unused Keys**: Disable or delete API keys that are no longer needed

## Examples

### Bash Script to Disable a Group

```bash
#!/bin/bash
API_KEY="your_api_key_here"
GROUP_NAME="VPN-Team-A"

curl -X POST "https://your-server.com/api/v1/group/set-status" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"group\":\"$GROUP_NAME\",\"enabled\":false}"
```

### Python Script to List Clients

```python
import requests

API_KEY = "your_api_key_here"
BASE_URL = "https://your-server.com"

headers = {
    "Authorization": f"Bearer {API_KEY}"
}

response = requests.get(f"{BASE_URL}/api/v1/clients", headers=headers)
clients = response.json()

for client in clients:
    print(f"{client['Client']['name']}: {client['Client']['group']}")
```

## Error Responses

All API endpoints return JSON responses with this structure:

**Success**:
```json
{
  "success": true,
  "message": "Operation completed successfully"
}
```

**Error**:
```json
{
  "success": false,
  "message": "Error description"
}
```

**Common HTTP Status Codes**:
- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid API key)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

## Troubleshooting

### API Key Not Working

1. Check that the API key is enabled in the management page
2. Verify the API key has the required permissions
3. Ensure you're using the correct Authorization header format
4. Check that the API key wasn't disabled or deleted

### Group Operations Not Working

1. Verify the group name is spelled correctly (case-sensitive)
2. Ensure the API key has `manage:groups` permission
3. Check that clients actually exist in that group

## Changelog

### Version 1.0 (Initial Release)
- Added API key management with permission system
- Implemented group-level enable/disable functionality
- Added API statistics and access logging
- Created REST API endpoints for client and group operations
