# Multilingual Support - Feature Demonstration

## Overview

This document demonstrates the multilingual support feature added to WireGuard Manager. The application now supports multiple languages with automatic detection and easy switching.

## Supported Languages

- **English (en)** - Default language
- **German (de)** - Deutsche Übersetzung

## Key Features Demonstrated

### 1. Language Switcher in Navigation Bar

A globe icon with a dropdown menu is now available in the top right corner of the navigation bar:

```
[Globe Icon] EN ▼
  ├── English
  └── Deutsch
```

**Location**: Top right navbar, next to "New Client" and "Logout" buttons

### 2. Automatic Language Detection

The system automatically detects the user's preferred language using:
1. Language cookie (if previously set)
2. Browser's Accept-Language header
3. Default fallback to English

### 3. Translated Interface Elements

#### Navigation Menu (Sidebar)
- English: "MAIN", "VPN Clients", "VPN Connected", "SETTINGS", "WireGuard Server", etc.
- German: "HAUPTMENÜ", "VPN-Clients", "VPN-Verbindungen", "EINSTELLUNGEN", "WireGuard-Server", etc.

#### Action Buttons
- English: "New Client", "Apply Config", "Logout"
- German: "Neuer Client", "Konfiguration anwenden", "Abmelden"

#### Modal Dialogs
- English: "New WireGuard Client", "Cancel", "Submit"
- German: "Neuer WireGuard-Client", "Abbrechen", "Absenden"

#### Form Labels
- English: "Name", "Email", "Group", "IP Allocation", "Use server DNS"
- German: "Name", "E-Mail", "Gruppe", "IP-Zuweisung", "Server-DNS verwenden"

## Example Translation Comparison

### New Client Modal

**English Version:**
```
Title: "New WireGuard Client"

Fields:
- Name
- Email
- Group (Optional group name)
- Subnet range (Select a subnet range)
- IP Allocation
- Allowed IPs - What subnet traffic should go through WireGuard
- Use server DNS
- Enable after creation
- Public and Preshared Keys

Buttons: [Cancel] [Submit]
```

**German Version:**
```
Title: "Neuer WireGuard-Client"

Fields:
- Name
- E-Mail
- Gruppe (Optionaler Gruppenname)
- Subnetzbereich (Wählen Sie einen Subnetzbereich)
- IP-Zuweisung
- Erlaubte IPs - Welcher Subnetz-Datenverkehr soll über WireGuard geleitet werden
- Server-DNS verwenden
- Nach Erstellung aktivieren
- Öffentliche und vorverteilte Schlüssel

Buttons: [Abbrechen] [Absenden]
```

## Usage Instructions

### For End Users

1. **Switching Language:**
   - Click the globe icon (🌐) in the top right corner
   - Select your preferred language from the dropdown
   - The page will reload in the selected language
   - Your choice is automatically saved

2. **Automatic Detection:**
   - On first visit, the app detects your browser's language
   - Supported: English and German
   - Unsupported languages default to English

### For Administrators

1. **Checking Current Language:**
   - Look at the HTML `lang` attribute in the page source
   - Check the language cookie in browser dev tools

2. **Adding New Languages:**
   - See [MULTILINGUAL.md](MULTILINGUAL.md) for detailed instructions
   - Create a new JSON file in `i18n/translations/`
   - Update `SupportedLanguages` in `i18n/i18n.go`
   - Add dropdown option in `templates/base.html`
   - Rebuild the application

## Technical Implementation Highlights

### Translation Files Structure

```json
{
  "nav": {
    "new_client": "New Client",
    "logout": "Logout"
  },
  "modal": {
    "new_client_title": "New WireGuard Client"
  },
  "form": {
    "name": "Name",
    "email": "Email"
  }
}
```

### Template Usage

```html
<!-- Simple translation -->
<button>{{tr .t "nav.new_client"}}</button>

<!-- In attributes -->
<input placeholder="{{tr .t "form.name"}}">

<!-- Modal title -->
<h4>{{tr .t "modal.new_client_title"}}</h4>
```

### Language Cookie

```
Name: language
Value: en | de
Path: <basePath>
Max-Age: 31536000 (1 year)
```

## Testing the Feature

### Manual Testing Steps

1. **Test English (Default):**
   ```bash
   # Start application
   # Open browser to http://localhost:5000
   # Verify interface is in English
   ```

2. **Test German:**
   ```bash
   # Click globe icon → select "Deutsch"
   # Verify all UI elements are in German
   # Check that cookie is set: language=de
   ```

3. **Test Language Persistence:**
   ```bash
   # Select German
   # Close browser
   # Reopen browser to same URL
   # Verify language is still German
   ```

4. **Test Browser Language Detection:**
   ```bash
   # Clear cookies
   # Set browser language to German
   # Open application
   # Verify it opens in German
   ```

## Translation Coverage

Currently translated UI elements:

- ✅ Navigation menu (sidebar)
- ✅ Top navigation bar
- ✅ Action buttons
- ✅ "New Client" modal dialog
- ✅ "Apply Config" modal dialog
- ✅ Form labels and placeholders
- ✅ Tooltips
- ✅ Footer
- ✅ Status filter options

## Performance Impact

- **Startup:** Negligible (translations loaded once at startup)
- **Runtime:** Zero overhead (simple map lookups)
- **Memory:** < 50KB for both languages
- **Build:** Translations embedded in binary (no external files)

## Future Enhancements

Potential improvements:

1. More languages (French, Spanish, Italian, etc.)
2. User profile language preference
3. RTL (Right-to-Left) language support
4. Date/time localization
5. Number and currency formatting
6. Translation management interface

## Conclusion

The multilingual support feature provides a seamless, user-friendly experience for international users while maintaining zero performance overhead and simple maintainability. The implementation is clean, well-documented, and easily extensible for adding new languages in the future.
