# WireGuard Manager API Dokumentation

Diese Dokumentation beschreibt die REST-API des WireGuard Managers, einschließlich API-Schlüssel-Verwaltung, Client-Operationen und Gruppen-Management.

## Inhaltsverzeichnis

1. [Übersicht](#übersicht)
2. [Schnellstart](#schnellstart)
3. [Authentifizierung](#authentifizierung)
4. [API-Endpunkte](#api-endpunkte)
   - [Client-Operationen](#client-operationen)
   - [Gruppen-Operationen](#gruppen-operationen)
5. [Berechtigungen](#berechtigungen)
6. [Fehlerbehandlung](#fehlerbehandlung)
7. [Nutzungsbeispiele](#nutzungsbeispiele)
8. [Best Practices](#best-practices)
9. [Fehlerbehebung](#fehlerbehebung)

## Übersicht

Die WireGuard Manager API bietet programmatischen Zugriff auf folgende Funktionen:

- **Client-Verwaltung**: Erstellen, Lesen, Aktualisieren und Löschen von WireGuard-Clients
- **Gruppen-Management**: Massenaktivierung/-deaktivierung von Clients in Gruppen
- **API-Schlüssel-Verwaltung**: Erstellen und Verwalten von API-Schlüsseln mit granularen Berechtigungen
- **API-Statistiken**: Verfolgung der API-Nutzung mit detaillierten Protokollen

## Schnellstart

### 1. API-Schlüssel erstellen

Navigieren Sie in der Web-Oberfläche zu **API → API-Schlüssel-Verwaltung**.

1. Klicken Sie auf "Neuer API-Schlüssel"
2. Geben Sie einen Namen für den Schlüssel ein
3. Wählen Sie die erforderlichen Berechtigungen aus
4. Klicken Sie auf "Erstellen"
5. **Wichtig**: Kopieren Sie den API-Schlüssel sofort - er wird nur einmal angezeigt!

### 2. Erster API-Aufruf

```bash
curl -X GET "https://ihr-server.de/api/v1/clients" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL"
```

## Authentifizierung

Alle externen API-Endpunkte befinden sich unter `/api/v1/` und erfordern eine Authentifizierung über Bearer Token.

### Authorization Header

Fügen Sie den API-Schlüssel in den Authorization-Header ein:

```
Authorization: Bearer IHR_API_SCHLÜSSEL
```

### Beispiel mit curl

```bash
curl -X GET "https://ihr-server.de/api/v1/clients" \
  -H "Authorization: Bearer a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"
```

## API-Endpunkte

### Client-Operationen

#### Alle Clients auflisten

Listet alle WireGuard-Clients auf.

**Endpunkt**: `GET /api/v1/clients`

**Erforderliche Berechtigung**: `read:clients`

**Anfrage**:
```bash
curl -X GET "https://ihr-server.de/api/v1/clients" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL"
```

**Antwort** (200 OK):
```json
[
  {
    "Client": {
      "id": "cm5xyz123abc",
      "name": "Client 1",
      "email": "client1@beispiel.de",
      "group": "GruppeA",
      "enabled": true,
      "allocated_ips": ["10.8.0.2/32"],
      "allowed_ips": ["0.0.0.0/0"],
      "extra_allowed_ips": [],
      "use_server_dns": true,
      "public_key": "AbCdEfGhIjKlMnOpQrStUvWxYz1234567890=",
      "preshared_key": "PreSharedKey1234567890AbCdEfGhIjKl=",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    "QRCode": ""
  }
]
```

#### Einzelnen Client abrufen

Ruft die Details eines einzelnen Clients ab.

**Endpunkt**: `GET /api/v1/client/:id`

**Erforderliche Berechtigung**: `read:clients`

**Anfrage**:
```bash
curl -X GET "https://ihr-server.de/api/v1/client/cm5xyz123abc" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL"
```

**Antwort** (200 OK):
```json
{
  "Client": {
    "id": "cm5xyz123abc",
    "name": "Client 1",
    "email": "client1@beispiel.de",
    "group": "GruppeA",
    "enabled": true,
    "allocated_ips": ["10.8.0.2/32"],
    "allowed_ips": ["0.0.0.0/0"],
    "extra_allowed_ips": [],
    "use_server_dns": true,
    "public_key": "AbCdEfGhIjKlMnOpQrStUvWxYz1234567890=",
    "preshared_key": "PreSharedKey1234567890AbCdEfGhIjKl=",
    "endpoint": "",
    "private_key": "PrivateKey1234567890AbCdEfGhIjKlMnOp=",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "QRCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
}
```

#### Client erstellen

Erstellt einen neuen WireGuard-Client.

**Endpunkt**: `POST /api/v1/client`

**Erforderliche Berechtigung**: `write:clients`

**Anfrage**:
```bash
curl -X POST "https://ihr-server.de/api/v1/client" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Neuer Client",
    "email": "neuer.client@beispiel.de",
    "group": "GruppeA",
    "allocated_ips": ["10.8.0.10/32"],
    "allowed_ips": ["0.0.0.0/0"],
    "extra_allowed_ips": [],
    "use_server_dns": true,
    "enabled": true
  }'
```

**Hinweise**:
- Wenn `public_key` nicht angegeben wird, wird automatisch ein Schlüsselpaar generiert
- Wenn `preshared_key` nicht angegeben wird, wird automatisch einer generiert
- Um die Generierung des Preshared-Keys zu überspringen, setzen Sie `preshared_key: "-"`
- Die `allocated_ips` müssen innerhalb der Server-Interface-Adressen liegen und dürfen nicht bereits vergeben sein

**Antwort** (200 OK):
```json
{
  "id": "cm5xyz789def",
  "name": "Neuer Client",
  "email": "neuer.client@beispiel.de",
  "group": "GruppeA",
  "enabled": true,
  "allocated_ips": ["10.8.0.10/32"],
  "allowed_ips": ["0.0.0.0/0"],
  "extra_allowed_ips": [],
  "use_server_dns": true,
  "public_key": "GeneratedPublicKey1234567890=",
  "private_key": "GeneratedPrivateKey1234567890=",
  "preshared_key": "GeneratedPresharedKey1234567890=",
  "created_at": "2024-01-20T14:25:00Z",
  "updated_at": "2024-01-20T14:25:00Z"
}
```

#### Client aktualisieren

Aktualisiert einen bestehenden WireGuard-Client.

**Endpunkt**: `PUT /api/v1/client`

**Erforderliche Berechtigung**: `write:clients`

**Anfrage**:
```bash
curl -X PUT "https://ihr-server.de/api/v1/client" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cm5xyz123abc",
    "name": "Aktualisierter Name",
    "email": "aktualisiert@beispiel.de",
    "group": "GruppeB",
    "enabled": true,
    "allocated_ips": ["10.8.0.2/32"],
    "allowed_ips": ["0.0.0.0/0", "192.168.1.0/24"],
    "extra_allowed_ips": [],
    "use_server_dns": true,
    "public_key": "ExistingPublicKey1234567890=",
    "preshared_key": "ExistingPresharedKey1234567890="
  }'
```

**Antwort** (200 OK):
```json
{
  "success": true,
  "message": "Client erfolgreich aktualisiert"
}
```

#### Client-Status setzen

Aktiviert oder deaktiviert einen einzelnen Client.

**Endpunkt**: `POST /api/v1/client/set-status`

**Erforderliche Berechtigung**: `write:clients`

**Anfrage**:
```bash
curl -X POST "https://ihr-server.de/api/v1/client/set-status" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cm5xyz123abc",
    "status": false
  }'
```

**Parameter**:
- `id`: Die Client-ID
- `status`: `true` zum Aktivieren, `false` zum Deaktivieren

**Antwort** (200 OK):
```json
{
  "success": true,
  "message": "Client-Status erfolgreich geändert"
}
```

#### Client löschen

Löscht einen WireGuard-Client permanent.

**Endpunkt**: `DELETE /api/v1/client/:id`

**Erforderliche Berechtigung**: `write:clients`

**Anfrage**:
```bash
curl -X DELETE "https://ihr-server.de/api/v1/client/cm5xyz123abc" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL" \
  -H "Content-Type: application/json"
```

**Antwort** (200 OK):
```json
{
  "success": true,
  "message": "Client entfernt"
}
```

### Gruppen-Operationen

#### Gruppenstatus setzen

Aktiviert oder deaktiviert alle Clients in einer Gruppe auf einmal.

**Endpunkt**: `POST /api/v1/group/set-status`

**Erforderliche Berechtigung**: `manage:groups`

**Anfrage**:
```bash
curl -X POST "https://ihr-server.de/api/v1/group/set-status" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL" \
  -H "Content-Type: application/json" \
  -d '{
    "group": "GruppeA",
    "enabled": false
  }'
```

**Parameter**:
- `group`: Name der Gruppe (Groß-/Kleinschreibung beachten)
- `enabled`: `true` zum Aktivieren, `false` zum Deaktivieren

**Antwort** (200 OK):
```json
{
  "success": true,
  "message": "5 Client(s) in Gruppe 'GruppeA' erfolgreich deaktiviert"
}
```

**Anwendungsfälle**:
- Temporäres Deaktivieren des Zugriffs für ein Team oder eine Abteilung
- Notabschaltung einer Gruppe von Clients
- Wartungsfenster

## Berechtigungen

API-Schlüssel unterstützen folgende Berechtigungen:

| Berechtigung | Beschreibung |
|--------------|--------------|
| `read:clients` | Anzeigen von Client-Informationen |
| `write:clients` | Erstellen, Aktualisieren und Löschen von Clients |
| `read:server` | Anzeigen der Server-Konfiguration |
| `write:server` | Ändern der Server-Konfiguration |
| `manage:groups` | Aktivieren/Deaktivieren von Gruppen |
| `read:stats` | Anzeigen von API-Statistiken |

**Best Practice**: Gewähren Sie nur die minimal erforderlichen Berechtigungen für jeden Anwendungsfall.

## Fehlerbehandlung

### Erfolgsantwort

```json
{
  "success": true,
  "message": "Operation erfolgreich abgeschlossen"
}
```

### Fehlerantwort

```json
{
  "success": false,
  "message": "Fehlerbeschreibung"
}
```

### HTTP-Statuscodes

| Code | Bedeutung |
|------|-----------|
| `200` | Erfolg |
| `400` | Ungültige Anfrage (fehlerhafte Eingabedaten) |
| `401` | Nicht authentifiziert (fehlender oder ungültiger API-Schlüssel) |
| `403` | Verboten (unzureichende Berechtigungen) |
| `404` | Nicht gefunden (Ressource existiert nicht) |
| `500` | Interner Serverfehler |

## Nutzungsbeispiele

### Bash/Shell-Skripte

#### Gruppe nach Zeitplan deaktivieren

```bash
#!/bin/bash
# Deaktiviert eine Gruppe außerhalb der Geschäftszeiten

API_KEY="ihr_api_schluessel_hier"
BASE_URL="https://ihr-server.de"
GROUP_NAME="Mitarbeiter-VPN"

# Gruppe deaktivieren
curl -X POST "$BASE_URL/api/v1/group/set-status" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"group\":\"$GROUP_NAME\",\"enabled\":false}"
```

#### Client mit automatischer IP-Zuweisung erstellen

```bash
#!/bin/bash

API_KEY="ihr_api_schluessel_hier"
BASE_URL="https://ihr-server.de"

# Clients aus CSV-Datei erstellen
# Format: name,email,group
while IFS=',' read -r name email group; do
  echo "Erstelle Client: $name"
  
  curl -X POST "$BASE_URL/api/v1/client" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"$name\",
      \"email\": \"$email\",
      \"group\": \"$group\",
      \"allocated_ips\": [\"10.8.0.0/24\"],
      \"allowed_ips\": [\"0.0.0.0/0\"],
      \"use_server_dns\": true,
      \"enabled\": true
    }"
  
  sleep 1
done < clients.csv
```

### Python-Beispiele

#### Alle Clients auflisten und filtern

```python
import requests
import json

API_KEY = "ihr_api_schluessel_hier"
BASE_URL = "https://ihr-server.de"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# Alle Clients abrufen
response = requests.get(f"{BASE_URL}/api/v1/clients", headers=headers)
clients = response.json()

# Nach Gruppe filtern
gruppe_a_clients = [
    client for client in clients 
    if client['Client']['group'] == 'GruppeA'
]

print(f"Clients in GruppeA: {len(gruppe_a_clients)}")
for client in gruppe_a_clients:
    c = client['Client']
    status = "Aktiv" if c['enabled'] else "Inaktiv"
    print(f"  - {c['name']} ({c['email']}): {status}")
```

#### Client mit Fehlerbehandlung erstellen

```python
import requests
import sys

API_KEY = "ihr_api_schluessel_hier"
BASE_URL = "https://ihr-server.de"

def create_client(name, email, group, ips):
    """Erstellt einen neuen WireGuard-Client mit Fehlerbehandlung"""
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "name": name,
        "email": email,
        "group": group,
        "allocated_ips": ips,
        "allowed_ips": ["0.0.0.0/0"],
        "use_server_dns": True,
        "enabled": True
    }
    
    try:
        response = requests.post(
            f"{BASE_URL}/api/v1/client",
            headers=headers,
            json=payload,
            timeout=30
        )
        
        if response.status_code == 200:
            client = response.json()
            print(f"✓ Client '{name}' erfolgreich erstellt")
            print(f"  ID: {client['id']}")
            print(f"  Öffentlicher Schlüssel: {client['public_key']}")
            return client
        else:
            error = response.json()
            print(f"✗ Fehler beim Erstellen des Clients: {error.get('message', 'Unbekannter Fehler')}")
            return None
            
    except requests.exceptions.Timeout:
        print("✗ Zeitüberschreitung bei der Verbindung zum Server")
        return None
    except requests.exceptions.RequestException as e:
        print(f"✗ Verbindungsfehler: {e}")
        return None

# Verwendung
if __name__ == "__main__":
    client = create_client(
        name="Test Client",
        email="test@beispiel.de",
        group="Test",
        ips=["10.8.0.50/32"]
    )
```

#### Batch-Operationen mit Fortschrittsanzeige

```python
import requests
from typing import List, Dict
import time

API_KEY = "ihr_api_schluessel_hier"
BASE_URL = "https://ihr-server.de"

class WireGuardAPIClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }
    
    def get_clients(self) -> List[Dict]:
        """Ruft alle Clients ab"""
        response = requests.get(
            f"{self.base_url}/api/v1/clients",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()
    
    def disable_inactive_clients(self, days_inactive: int = 30) -> int:
        """Deaktiviert Clients, die seit X Tagen inaktiv sind"""
        from datetime import datetime, timedelta
        
        clients = self.get_clients()
        cutoff_date = datetime.utcnow() - timedelta(days=days_inactive)
        disabled_count = 0
        
        for client_data in clients:
            client = client_data['Client']
            if not client['enabled']:
                continue
                
            updated_at = datetime.fromisoformat(
                client['updated_at'].replace('Z', '+00:00')
            )
            
            if updated_at < cutoff_date:
                print(f"Deaktiviere inaktiven Client: {client['name']}")
                response = requests.post(
                    f"{self.base_url}/api/v1/client/set-status",
                    headers=self.headers,
                    json={"id": client['id'], "status": False}
                )
                
                if response.status_code == 200:
                    disabled_count += 1
                    time.sleep(0.5)  # Rate limiting
        
        return disabled_count

# Verwendung
api = WireGuardAPIClient(BASE_URL, API_KEY)
count = api.disable_inactive_clients(days_inactive=60)
print(f"Insgesamt {count} inaktive Clients deaktiviert")
```

### JavaScript/Node.js-Beispiele

#### Einfache Client-Abfrage

```javascript
const axios = require('axios');

const API_KEY = 'ihr_api_schluessel_hier';
const BASE_URL = 'https://ihr-server.de';

const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Authorization': `Bearer ${API_KEY}`,
    'Content-Type': 'application/json'
  }
});

// Alle Clients abrufen
async function getAllClients() {
  try {
    const response = await api.get('/api/v1/clients');
    return response.data;
  } catch (error) {
    console.error('Fehler beim Abrufen der Clients:', error.response?.data || error.message);
    throw error;
  }
}

// Client erstellen
async function createClient(name, email, group, allocatedIPs) {
  try {
    const response = await api.post('/api/v1/client', {
      name,
      email,
      group,
      allocated_ips: allocatedIPs,
      allowed_ips: ['0.0.0.0/0'],
      use_server_dns: true,
      enabled: true
    });
    
    console.log(`Client '${name}' erfolgreich erstellt:`, response.data.id);
    return response.data;
  } catch (error) {
    console.error('Fehler beim Erstellen des Clients:', error.response?.data || error.message);
    throw error;
  }
}

// Verwendung
(async () => {
  const clients = await getAllClients();
  console.log(`Insgesamt ${clients.length} Clients`);
  
  // Neuen Client erstellen
  await createClient(
    'Neuer JS Client',
    'jsuser@beispiel.de',
    'Entwickler',
    ['10.8.0.100/32']
  );
})();
```

#### Express.js-Middleware für Proxy

```javascript
const express = require('express');
const axios = require('axios');

const app = express();
app.use(express.json());

const WIREGUARD_API_KEY = process.env.WIREGUARD_API_KEY;
const WIREGUARD_BASE_URL = process.env.WIREGUARD_BASE_URL;

// Middleware für WireGuard API-Aufrufe
app.use('/wireguard', async (req, res) => {
  try {
    const response = await axios({
      method: req.method,
      url: `${WIREGUARD_BASE_URL}/api/v1${req.path}`,
      headers: {
        'Authorization': `Bearer ${WIREGUARD_API_KEY}`,
        'Content-Type': 'application/json'
      },
      data: req.body
    });
    
    res.json(response.data);
  } catch (error) {
    res.status(error.response?.status || 500).json({
      success: false,
      message: error.response?.data?.message || error.message
    });
  }
});

app.listen(3000, () => {
  console.log('Proxy läuft auf Port 3000');
});
```

### Go-Beispiele

#### Vollständiger API-Client

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	APIKey  = "ihr_api_schluessel_hier"
	BaseURL = "https://ihr-server.de"
)

type Client struct {
	ID            string    `json:"id,omitempty"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Group         string    `json:"group"`
	Enabled       bool      `json:"enabled"`
	AllocatedIPs  []string  `json:"allocated_ips"`
	AllowedIPs    []string  `json:"allowed_ips"`
	UseServerDNS  bool      `json:"use_server_dns"`
	PublicKey     string    `json:"public_key,omitempty"`
	PrivateKey    string    `json:"private_key,omitempty"`
	PresharedKey  string    `json:"preshared_key,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type ClientData struct {
	Client Client `json:"Client"`
	QRCode string `json:"QRCode"`
}

type WireGuardAPI struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewWireGuardAPI(baseURL, apiKey string) *WireGuardAPI {
	return &WireGuardAPI{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (api *WireGuardAPI) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, api.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+api.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return api.client.Do(req)
}

func (api *WireGuardAPI) GetClients() ([]ClientData, error) {
	resp, err := api.doRequest("GET", "/api/v1/clients", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API-Fehler: %s", string(body))
	}

	var clients []ClientData
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (api *WireGuardAPI) CreateClient(client Client) (*Client, error) {
	resp, err := api.doRequest("POST", "/api/v1/client", client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API-Fehler: %s", string(body))
	}

	var newClient Client
	if err := json.NewDecoder(resp.Body).Decode(&newClient); err != nil {
		return nil, err
	}

	return &newClient, nil
}

func (api *WireGuardAPI) SetGroupStatus(group string, enabled bool) error {
	payload := map[string]interface{}{
		"group":   group,
		"enabled": enabled,
	}

	resp, err := api.doRequest("POST", "/api/v1/group/set-status", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API-Fehler: %s", string(body))
	}

	return nil
}

func main() {
	api := NewWireGuardAPI(BaseURL, APIKey)

	// Alle Clients abrufen
	clients, err := api.GetClients()
	if err != nil {
		fmt.Printf("Fehler beim Abrufen der Clients: %v\n", err)
		return
	}

	fmt.Printf("Insgesamt %d Clients gefunden\n", len(clients))
	for _, clientData := range clients {
		c := clientData.Client
		status := "Aktiv"
		if !c.Enabled {
			status = "Inaktiv"
		}
		fmt.Printf("  - %s (%s): %s - Gruppe: %s\n", c.Name, c.Email, status, c.Group)
	}

	// Neuen Client erstellen
	newClient := Client{
		Name:         "Go Test Client",
		Email:        "gotest@beispiel.de",
		Group:        "Test",
		AllocatedIPs: []string{"10.8.0.150/32"},
		AllowedIPs:   []string{"0.0.0.0/0"},
		UseServerDNS: true,
		Enabled:      true,
	}

	created, err := api.CreateClient(newClient)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen des Clients: %v\n", err)
		return
	}

	fmt.Printf("\nNeuer Client erstellt:\n")
	fmt.Printf("  ID: %s\n", created.ID)
	fmt.Printf("  Öffentlicher Schlüssel: %s\n", created.PublicKey)
}
```

### PowerShell-Beispiele

#### Client-Verwaltung mit PowerShell

```powershell
# WireGuard Manager API Client für PowerShell

$ApiKey = "ihr_api_schluessel_hier"
$BaseUrl = "https://ihr-server.de"

$Headers = @{
    "Authorization" = "Bearer $ApiKey"
    "Content-Type" = "application/json"
}

# Funktion: Alle Clients abrufen
function Get-WGClients {
    $Response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/clients" -Headers $Headers -Method Get
    return $Response
}

# Funktion: Client erstellen
function New-WGClient {
    param(
        [string]$Name,
        [string]$Email,
        [string]$Group,
        [string[]]$AllocatedIPs
    )
    
    $Body = @{
        name = $Name
        email = $Email
        group = $Group
        allocated_ips = $AllocatedIPs
        allowed_ips = @("0.0.0.0/0")
        use_server_dns = $true
        enabled = $true
    } | ConvertTo-Json
    
    try {
        $Response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/client" -Headers $Headers -Method Post -Body $Body
        Write-Host "✓ Client '$Name' erfolgreich erstellt (ID: $($Response.id))" -ForegroundColor Green
        return $Response
    } catch {
        Write-Host "✗ Fehler beim Erstellen des Clients: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Funktion: Gruppenstatus setzen
function Set-WGGroupStatus {
    param(
        [string]$Group,
        [bool]$Enabled
    )
    
    $Body = @{
        group = $Group
        enabled = $Enabled
    } | ConvertTo-Json
    
    try {
        $Response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/group/set-status" -Headers $Headers -Method Post -Body $Body
        $Action = if ($Enabled) { "aktiviert" } else { "deaktiviert" }
        Write-Host "✓ Gruppe '$Group' $Action" -ForegroundColor Green
        return $Response
    } catch {
        Write-Host "✗ Fehler: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Verwendungsbeispiele

# Alle Clients anzeigen
$Clients = Get-WGClients
Write-Host "Insgesamt $($Clients.Count) Clients gefunden"

# Clients nach Gruppe gruppieren
$Clients | Group-Object { $_.Client.group } | ForEach-Object {
    Write-Host "`n$($_.Name): $($_.Count) Clients"
    $_.Group | ForEach-Object {
        $Status = if ($_.Client.enabled) { "Aktiv" } else { "Inaktiv" }
        Write-Host "  - $($_.Client.name) ($Status)"
    }
}

# Neuen Client erstellen
New-WGClient -Name "PS Test Client" -Email "pstest@beispiel.de" -Group "Test" -AllocatedIPs @("10.8.0.200/32")

# Gruppe deaktivieren
Set-WGGroupStatus -Group "Test" -Enabled $false
```

## Best Practices

### Sicherheit

1. **API-Schlüssel sicher aufbewahren**
   - Niemals API-Schlüssel in Versionskontrolle committen
   - Verwenden Sie Umgebungsvariablen oder sichere Secrets-Manager
   - Rotieren Sie Schlüssel regelmäßig

2. **Minimale Berechtigungen verwenden**
   - Vergeben Sie nur die benötigten Berechtigungen
   - Erstellen Sie separate Schlüssel für unterschiedliche Anwendungsfälle
   - Dokumentieren Sie den Zweck jedes Schlüssels

3. **Überwachen Sie die API-Nutzung**
   - Prüfen Sie regelmäßig die API-Statistiken
   - Achten Sie auf ungewöhnliche Aktivitäten
   - Deaktivieren oder löschen Sie ungenutzte Schlüssel

### Performance

1. **Rate Limiting beachten**
   - Implementieren Sie Verzögerungen zwischen Batch-Operationen
   - Verwenden Sie Exponential Backoff bei Fehlern
   - Cachen Sie Daten, wenn möglich

2. **Batch-Operationen optimieren**
   ```python
   # Gut: Mit Verzögerung
   for client in clients:
       api.update_client(client)
       time.sleep(0.5)
   
   # Schlecht: Ohne Verzögerung (kann Server überlasten)
   for client in clients:
       api.update_client(client)
   ```

3. **Fehlerbehandlung**
   - Implementieren Sie Retry-Logik für transiente Fehler
   - Loggen Sie Fehler für Debugging
   - Verwenden Sie Timeouts für alle Requests

### Wartbarkeit

1. **API-Client abstrahieren**
   - Erstellen Sie eine Client-Bibliothek für Ihre Sprache
   - Zentralisieren Sie API-Aufrufe
   - Versionieren Sie Ihre Client-Bibliothek

2. **Konfiguration externalisieren**
   ```python
   # Gut: Aus Umgebungsvariablen
   API_KEY = os.getenv('WIREGUARD_API_KEY')
   BASE_URL = os.getenv('WIREGUARD_BASE_URL')
   
   # Schlecht: Hardcodiert
   API_KEY = "a1b2c3d4e5f6..."
   BASE_URL = "https://meinserver.de"
   ```

3. **Logging implementieren**
   ```python
   import logging
   
   logger = logging.getLogger(__name__)
   logger.info(f"Client erstellt: {client_id}")
   logger.error(f"Fehler beim Erstellen: {error}")
   ```

## Fehlerbehebung

### API-Schlüssel funktioniert nicht

**Problem**: API-Aufrufe werden mit 401 Unauthorized abgelehnt.

**Lösungen**:
1. Überprüfen Sie, ob der API-Schlüssel in der Verwaltungsseite aktiviert ist
2. Stellen Sie sicher, dass das Authorization-Header-Format korrekt ist: `Bearer <key>`
3. Prüfen Sie, ob der Schlüssel nicht gelöscht wurde
4. Verifizieren Sie, dass keine Leerzeichen im Schlüssel sind

**Test**:
```bash
# Schlüssel testen
curl -v -X GET "https://ihr-server.de/api/v1/clients" \
  -H "Authorization: Bearer IHR_API_SCHLÜSSEL"

# Überprüfen Sie die Antwort auf Details
```

### Unzureichende Berechtigungen

**Problem**: API-Aufrufe werden mit 403 Forbidden abgelehnt.

**Lösungen**:
1. Überprüfen Sie die Berechtigungen des API-Schlüssels in der Web-Oberfläche
2. Stellen Sie sicher, dass die erforderliche Berechtigung zugewiesen ist
3. Für `write:clients` Operationen benötigen Sie die entsprechende Berechtigung

**Beispiel**:
```bash
# Für diesen Aufruf benötigen Sie "write:clients"
curl -X POST "https://ihr-server.de/api/v1/client" ...
```

### IP-Allokation schlägt fehl

**Problem**: Beim Erstellen eines Clients erscheint "IP already allocated" oder "Invalid IP allocation".

**Lösungen**:
1. Überprüfen Sie, ob die IP-Adresse bereits einem anderen Client zugewiesen ist
2. Stellen Sie sicher, dass die IP innerhalb der Server-Interface-Adressen liegt
3. Verwenden Sie das korrekte CIDR-Format (z.B. `10.8.0.50/32` für einzelne IP)
4. Lassen Sie das System eine verfügbare IP automatisch zuweisen (falls unterstützt)

**Beispiel**:
```json
{
  "allocated_ips": ["10.8.0.50/32"],  // Richtig: CIDR-Notation
  // NICHT: "allocated_ips": ["10.8.0.50"]
}
```

### Verbindungsprobleme

**Problem**: Timeouts oder Verbindungsfehler.

**Lösungen**:
1. Überprüfen Sie die Netzwerkverbindung zum Server
2. Stellen Sie sicher, dass die Firewall Port 5000 (oder Ihr konfigurierter Port) erlaubt
3. Verwenden Sie HTTPS für Produktionsumgebungen
4. Erhöhen Sie den Timeout-Wert in Ihrem Code

**Python-Beispiel**:
```python
response = requests.get(
    url,
    headers=headers,
    timeout=30  # 30 Sekunden Timeout
)
```

### Ungültige JSON-Daten

**Problem**: 400 Bad Request mit "Invalid client data" oder ähnlichen Fehlern.

**Lösungen**:
1. Validieren Sie Ihre JSON-Daten vor dem Senden
2. Stellen Sie sicher, dass alle erforderlichen Felder vorhanden sind
3. Überprüfen Sie die Datentypen (z.B. `boolean` für `enabled`, `array` für IPs)
4. Verwenden Sie JSON-Validatoren zur Fehlersuche

**Beispiel**:
```python
import json

# Daten vor dem Senden validieren
try:
    json_data = json.dumps(data)
    print("JSON gültig")
except ValueError as e:
    print(f"Ungültiges JSON: {e}")
```

## Erweiterte Themen

### Webhooks (zukünftig)

Obwohl aktuell nicht unterstützt, können Sie Polling verwenden, um auf Änderungen zu reagieren:

```python
import time

def poll_for_changes(api, interval=60):
    """Prüft alle 60 Sekunden auf Änderungen"""
    last_update = None
    
    while True:
        clients = api.get_clients()
        latest_update = max(
            c['Client']['updated_at'] for c in clients
        )
        
        if last_update and latest_update > last_update:
            print("Änderung erkannt!")
            # Ihre Logik hier
        
        last_update = latest_update
        time.sleep(interval)
```

### Automatisierung mit Cron

Beispiel für automatische Wartungsaufgaben:

```bash
# /etc/cron.d/wireguard-maintenance

# Deaktiviert inaktive Clients täglich um 2 Uhr nachts
0 2 * * * /usr/local/bin/disable-inactive-clients.sh

# Backup der Client-Liste täglich um 3 Uhr nachts
0 3 * * * /usr/local/bin/backup-clients.sh
```

### Integration in CI/CD

GitHub Actions Beispiel:

```yaml
name: WireGuard Client Deployment

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Create VPN Client
        env:
          WIREGUARD_API_KEY: ${{ secrets.WIREGUARD_API_KEY }}
          WIREGUARD_URL: ${{ secrets.WIREGUARD_URL }}
        run: |
          curl -X POST "$WIREGUARD_URL/api/v1/client" \
            -H "Authorization: Bearer $WIREGUARD_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
              "name": "CI-Runner-${{ github.run_id }}",
              "email": "ci@beispiel.de",
              "group": "CI-CD",
              "allocated_ips": ["10.8.0.0/24"],
              "allowed_ips": ["0.0.0.0/0"],
              "enabled": true
            }'
```

## Zusammenfassung

Die WireGuard Manager API bietet umfassende Möglichkeiten zur programmatischen Verwaltung Ihrer VPN-Infrastruktur. Mit den richtigen Berechtigungen und Best Practices können Sie:

- **Automatisieren** Sie Client-Bereitstellung und -Verwaltung
- **Integrieren** Sie VPN-Management in Ihre bestehenden Systeme
- **Skalieren** Sie Ihre VPN-Infrastruktur effizient
- **Überwachen** Sie die API-Nutzung und -Aktivitäten

Für weitere Fragen oder Unterstützung besuchen Sie die [GitHub-Repository](https://github.com/swissmakers/wireguard-manager) oder erstellen Sie ein Issue.

---

**Version**: 1.0  
**Letzte Aktualisierung**: Januar 2026  
**Sprache**: Deutsch
