# Schnellstart-Anleitung für Entwicklung

Diese Anleitung beschreibt die neu hinzugefügten Entwicklungswerkzeuge für das WireGuard Manager Projekt.

## Neu hinzugefügt

### 1. JetBrains GoLand / IntelliJ IDEA Projektkonfiguration

Das Projekt enthält jetzt vollständige GoLand/IntelliJ IDEA Konfigurationsdateien im `.idea` Verzeichnis.

#### Verfügbare Konfigurationen:

1. **Prepare Assets** - Bereitet Frontend-Assets vor (npm install, Kopieren von Dateien)
2. **Build WireGuard Manager** - Erstellt die Anwendung mit Entwicklungs-Flags
3. **Run WireGuard Manager** - Startet die Anwendung mit Standard-Umgebungsvariablen
4. **Debug WireGuard Manager** - Startet die Anwendung im Debug-Modus mit Breakpoint-Unterstützung

#### Verwendung:

1. Öffnen Sie GoLand/IntelliJ IDEA
2. Öffnen Sie das Projektverzeichnis
3. Die Konfigurationen erscheinen automatisch im Run-Menü
4. Wählen Sie "Run WireGuard Manager" oder "Debug WireGuard Manager"

### 2. Jenkins CI/CD Pipeline

Eine neue `Jenkinsfile` wurde hinzugefügt, die automatisches Bauen auf Jenkins ermöglicht.

#### Features:

- **Multi-Plattform-Builds**: Linux x64 und Windows x64 werden parallel gebaut
- **Asset-Vorbereitung**: Automatische Vorbereitung der Frontend-Assets
- **Tests**: Automatische Ausführung von Go-Tests
- **Artefakte**: Fertige Binärdateien werden archiviert und können heruntergeladen werden

#### Build-Ausgaben:

- `wireguard-manager-linux-amd64` - Linux x64 ausführbare Datei
- `wireguard-manager-windows-amd64.exe` - Windows x64 ausführbare Datei

#### Jenkins-Voraussetzungen:

- Jenkins 2.x oder neuer
- Go 1.23+ installiert
- Node.js 18+ installiert
- Linux-Agent mit Label "linux"
- Optional: Windows-Agent mit Label "windows"

## Weitere Informationen

Ausführliche Dokumentation finden Sie in [DEVELOPMENT.md](DEVELOPMENT.md).

## Manuelle Builds

Wenn Sie ohne IDE oder CI/CD bauen möchten:

### Assets vorbereiten:
```bash
chmod +x ./prepare_assets.sh
./prepare_assets.sh
```

### Für Linux x64 bauen:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags="-X 'main.appVersion=dev' -X 'main.buildTime=$(date)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.gitRef=$(git branch --show-current)'" \
  -o wireguard-manager-linux-amd64 \
  .
```

### Für Windows x64 bauen:
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
  -ldflags="-X 'main.appVersion=dev' -X 'main.buildTime=$(date)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.gitRef=$(git branch --show-current)'" \
  -o wireguard-manager-windows-amd64.exe \
  .
```

## Problembehandlung

### Assets nicht gefunden
Führen Sie zuerst "Prepare Assets" aus:
```bash
./prepare_assets.sh
```

### GoLand erkennt Go-Projekt nicht
1. Stellen Sie sicher, dass Go SDK konfiguriert ist: Datei → Einstellungen → Go → GOROOT
2. Aktivieren Sie Go-Module: Datei → Einstellungen → Go → Go-Module → Go-Module-Integration aktivieren

### Jenkins-Build schlägt bei Asset-Vorbereitung fehl
1. Stellen Sie sicher, dass Node.js auf dem Jenkins-Agent installiert ist
2. Überprüfen Sie, dass npm/yarn im PATH zugänglich ist
3. Überprüfen Sie den Netzwerkzugriff zum Herunterladen von Node-Modulen
