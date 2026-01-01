# Security Features Documentation

This document describes the new security features added to the WireGuard Manager application.

## Overview

The application now includes comprehensive security measures to protect against various types of attacks:

1. **Brute Force Protection** - Limits failed login attempts from a single IP address
2. **IP Blocking** - Manual and automatic IP address blocking
3. **GeoIP Blocking** - Country-based access control (framework ready, requires GeoIP service integration)

## Features

### 1. Brute Force Protection

Automatically blocks IP addresses that exceed a configurable number of failed login attempts within a specified time window.

**Configuration:**
- **Enable/Disable**: Toggle brute force protection on/off
- **Max Attempts**: Number of failed login attempts before blocking (default: 5)
- **Time Window**: Time period in minutes to count failed attempts (default: 15)
- **Block Duration**: How long to block the IP in minutes (default: 30)

**How it works:**
- Each failed login attempt from an IP is recorded
- If the IP exceeds the maximum attempts within the time window, it's temporarily blocked
- After the block duration expires, the IP can attempt to login again
- Successful login clears the failed attempt counter

### 2. IP Blocking

Administrators can manually block IP addresses or view automatically blocked IPs.

**Features:**
- Manual IP blocking with optional reason
- Permanent or temporary blocks (with configurable expiration)
- View all blocked IPs with block reasons and expiration times
- Quick unblock functionality

**Use cases:**
- Block known malicious IP addresses
- Block IPs from security event statistics
- Temporary blocks for investigation

### 3. GeoIP Blocking

Framework for country-based access control (requires external GeoIP service integration).

**Configuration:**
- **Enable/Disable**: Toggle GeoIP blocking on/off
- **Default Action**: Allow or block countries without specific rules
- **Country Rules**: Create allow/block rules for specific countries

**Note:** This feature requires integration with a GeoIP lookup service to identify the country of incoming requests. The database schema and UI are ready, but the actual GeoIP lookup needs to be implemented based on your preferred GeoIP provider.

## Proxy Support

The application correctly handles IP addresses when running behind a proxy server (e.g., nginx, Apache).

**Supported Headers:**
- `X-Forwarded-For`: Multiple proxy chain support
- `X-Real-IP`: Single proxy support

**Configuration:**
Set the `PROXY` environment variable or use the `--proxy` flag to enable proxy mode:
```bash
export PROXY=true
./wireguard-manager
```

Or with Docker:
```bash
docker run -e PROXY=true ...
```

## Admin Pages

### Security Settings

Navigate to **Security Settings** from the admin menu to configure security features:

- Configure brute force protection settings
- View and manage IP blocks
- Configure GeoIP rules

**Path:** `/security-settings`

### Security Statistics

Navigate to **Security Statistics** to view security events and statistics:

- Total security events count
- Failed login attempts
- Blocked IPs count
- Brute force blocks count
- Recent security events timeline
- Top offending IP addresses
- Events breakdown by type

**Path:** `/security-statistics`

## API Endpoints

All security-related API endpoints require admin authentication:

### Settings
- `GET /api/security/settings` - Get security settings
- `POST /api/security/settings` - Update security settings

### Events
- `GET /api/security/events` - Get security events (last 100)
- `GET /api/security/statistics` - Get aggregated statistics

### IP Blocks
- `GET /api/security/ip-blocks` - Get all IP blocks
- `POST /api/security/ip-blocks` - Create IP block
- `DELETE /api/security/ip-blocks` - Remove IP block

### GeoIP Rules
- `GET /api/security/geoip-rules` - Get all GeoIP rules
- `POST /api/security/geoip-rules` - Create GeoIP rule
- `DELETE /api/security/geoip-rules` - Remove GeoIP rule

## Database Support

Security features are supported on both database backends:

- **JSON Database**: All security data stored in separate collections under the database path
- **MySQL Database**: Dedicated tables for security settings, events, IP blocks, GeoIP rules, and brute force attempts

### MySQL Schema

The following tables are automatically created on initialization:

- `security_settings` - Security configuration
- `security_events` - Security event log
- `ip_blocks` - Blocked IP addresses
- `geoip_rules` - Country-based rules
- `brute_force_attempts` - Failed login tracking

## Security Events

The system logs various types of security events:

- `failed_login` - Failed login attempt
- `blocked_ip` - Access attempt from blocked IP
- `blocked_geoip` - Access attempt from blocked country
- `brute_force` - Brute force block triggered

Events include:
- Timestamp
- Event type
- IP address
- Username (when applicable)
- Description
- Country (when GeoIP is enabled)

## Best Practices

1. **Enable Brute Force Protection**: Always keep this enabled in production
2. **Monitor Statistics**: Regularly check the security statistics page for suspicious activity
3. **Review Events**: Investigate repeated failed login attempts
4. **Use Permanent Blocks Carefully**: Only use permanent IP blocks for confirmed malicious actors
5. **Behind a Proxy**: Always set `PROXY=true` when running behind a reverse proxy
6. **Regular Cleanup**: The system automatically cleans up expired brute force attempts

## Limitations

1. **GeoIP Lookup**: Requires integration with external GeoIP service (MaxMind, IP2Location, etc.)
2. **IPv6 Support**: Full IPv6 addresses are supported but may need additional validation depending on your network setup
3. **Performance**: For high-traffic deployments, consider using MySQL instead of JSON database for better performance

## Future Enhancements

Potential improvements for future versions:

- Integration with popular GeoIP services
- Rate limiting per user account
- Two-factor authentication (2FA)
- CAPTCHA integration after multiple failed attempts
- Security event alerting via email/webhook
- Export security reports
- Automated IP reputation checking
