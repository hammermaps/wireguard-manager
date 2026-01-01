#!/usr/bin/env node
/**
 * List all WireGuard clients
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

async function listClients() {
  try {
    const response = await api.get('/api/v1/clients');
    const clients = response.data;
    
    console.log(`Total clients: ${clients.length}\n`);
    
    // Count enabled/disabled
    const enabled = clients.filter(c => c.Client.enabled).length;
    const disabled = clients.length - enabled;
    console.log(`Enabled: ${enabled}`);
    console.log(`Disabled: ${disabled}\n`);
    
    // Group by group name
    const groups = {};
    clients.forEach(clientData => {
      const client = clientData.Client;
      const group = client.group || 'No Group';
      if (!groups[group]) {
        groups[group] = [];
      }
      groups[group].push(client);
    });
    
    // Print by group
    Object.keys(groups).sort().forEach(group => {
      const groupClients = groups[group];
      console.log(`\n${group}: ${groupClients.length} clients`);
      groupClients.forEach(client => {
        const status = client.enabled ? '✓' : '✗';
        const ips = client.allocated_ips.join(', ');
        console.log(`  ${status} ${client.name} (${client.email}) - ${ips}`);
      });
    });
    
  } catch (error) {
    if (error.response) {
      console.error('API Error:', error.response.data);
    } else {
      console.error('Error:', error.message);
    }
    process.exit(1);
  }
}

// Run
listClients();
