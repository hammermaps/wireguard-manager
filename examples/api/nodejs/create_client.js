#!/usr/bin/env node
/**
 * Create a new WireGuard client
 */

const axios = require('axios');

// Get API configuration from environment
const API_KEY = process.env.WIREGUARD_API_KEY;
const BASE_URL = process.env.WIREGUARD_BASE_URL;

if (!API_KEY) {
  console.error('Error: WIREGUARD_API_KEY environment variable not set');
  process.exit(1);
}

if (!BASE_URL) {
  console.error('Error: WIREGUARD_BASE_URL environment variable not set');
  process.exit(1);
}

// Create API client
const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Authorization': `Bearer ${API_KEY}`,
    'Content-Type': 'application/json'
  },
  timeout: 30000
});

async function createClient(name, email, group, ipRange = '10.8.0.0/24') {
  try {
    const response = await api.post('/api/v1/client', {
      name,
      email,
      group,
      allocated_ips: [ipRange],
      allowed_ips: ['0.0.0.0/0'],
      use_server_dns: true,
      enabled: true
    });
    
    const client = response.data;
    console.log(`✓ Client '${name}' created successfully!`);
    console.log(`  ID: ${client.id}`);
    console.log(`  Public Key: ${client.public_key}`);
    console.log(`  Allocated IPs: ${client.allocated_ips.join(', ')}`);
    
    return client;
    
  } catch (error) {
    if (error.response) {
      console.error('✗ Error creating client:', error.response.data.message || error.response.data);
    } else {
      console.error('✗ Connection error:', error.message);
    }
    process.exit(1);
  }
}

// Parse command line arguments
const args = process.argv.slice(2);
if (args.length < 3) {
  console.log('Usage: node create_client.js <name> <email> <group> [ip_range]');
  console.log('Example: node create_client.js "John Doe" "john@example.com" "Employees" "10.8.0.50/32"');
  process.exit(1);
}

const [name, email, group, ipRange] = args;

// Run
createClient(name, email, group, ipRange);
