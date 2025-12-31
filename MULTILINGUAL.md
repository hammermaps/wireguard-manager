# Multilingual Support Documentation

## Overview

The WireGuard Manager now supports multiple languages, allowing users to view the interface in their preferred language. Currently supported languages are:

- **English (en)** - Default language
- **German (de)** - Deutsche Übersetzung

## Features

### Automatic Language Detection

The application automatically detects the user's preferred language using:
1. **Language Cookie**: If a user has previously selected a language, it will be stored in a cookie
2. **Accept-Language Header**: If no cookie is set, the application reads the browser's Accept-Language header
3. **Default Fallback**: English is used as the default language if no preference is detected

### Language Switcher

A language switcher is available in the top navigation bar:
- Click the globe icon to see available languages
- Select your preferred language from the dropdown menu
- The page will reload in the selected language
- Your choice is saved in a cookie for future visits

## Translation System

### Translation Files

Translation files are stored in JSON format in the `i18n/translations/` directory:

- `en.json` - English translations
- `de.json` - German translations

### Translation Structure

Translations are organized hierarchically by section:

```json
{
  "nav": {
    "vpn_clients": "VPN Clients",
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

### Using Translations in Templates

Translations are accessed using the `tr` template function:

```html
<!-- Simple translation -->
<h1>{{tr .t "page.vpn_clients_title"}}</h1>

<!-- Translation in attributes -->
<input placeholder="{{tr .t "nav.search_placeholder"}}">

<!-- Nested translation keys -->
<label>{{tr .t "form.email"}}</label>
```

## Adding New Languages

To add support for a new language:

1. **Create Translation File**:
   - Create a new JSON file in `i18n/translations/` (e.g., `fr.json` for French)
   - Copy the structure from `en.json`
   - Translate all text values to the new language

2. **Register Language**:
   - Edit `i18n/i18n.go`
   - Add the language code to the `SupportedLanguages` array:
     ```go
     var SupportedLanguages = []string{"en", "de", "fr"}
     ```

3. **Update Language Switcher**:
   - Edit `templates/base.html`
   - Add the new language option in the dropdown:
     ```html
     <a class="dropdown-item" href="{{.basePath}}/set-language?lang=fr">Français</a>
     ```

4. **Rebuild Application**:
   ```bash
   go build
   ```

## Technical Implementation

### Components

1. **i18n Package** (`i18n/i18n.go`):
   - Loads translation files at startup
   - Provides helper functions for accessing translations
   - Handles language detection from HTTP headers

2. **Router Middleware** (`router/router.go`):
   - Detects user's language preference
   - Injects translation data into template context
   - Provides the `tr` template function for accessing translations

3. **Language Handler** (`handler/routes.go`):
   - `/set-language?lang=XX` endpoint
   - Sets language cookie
   - Redirects back to previous page

4. **Templates** (`templates/base.html`):
   - Updated to use translation keys instead of hardcoded text
   - Language switcher in navigation bar
   - Dynamic `lang` attribute on HTML tag

### Language Cookie

The language preference is stored in a cookie:
- **Name**: `language`
- **Value**: Language code (e.g., `en`, `de`)
- **Path**: Application base path
- **Max-Age**: 1 year (31,536,000 seconds)
- **HttpOnly**: `false` (allows JavaScript access if needed)
- **SameSite**: `Lax`

## Environment Variables

No additional environment variables are required for multilingual support. The feature works out of the box with the default configuration.

## Performance Considerations

- Translation files are loaded once at application startup and cached in memory
- No database queries are required for translations
- Minimal overhead on page rendering (simple map lookups)
- Translation files are embedded in the binary (no external file dependencies)

## Extending Translations

To add new translatable strings:

1. **Add to Translation Files**:
   - Add the new key-value pair to all translation files
   - Use hierarchical keys for organization (e.g., `page.new_section.title`)

2. **Update Templates**:
   - Use the `tr` function to access the new translation
   - Example: `{{tr .t "page.new_section.title"}}`

3. **Maintain Consistency**:
   - Keep the same key structure across all language files
   - Provide translations for all supported languages
   - Use clear, descriptive key names

## Example: Adding a New Translatable String

1. Add to `i18n/translations/en.json`:
```json
{
  "page": {
    "welcome_message": "Welcome to WireGuard Manager!"
  }
}
```

2. Add to `i18n/translations/de.json`:
```json
{
  "page": {
    "welcome_message": "Willkommen beim WireGuard Manager!"
  }
}
```

3. Use in template:
```html
<h1>{{tr .t "page.welcome_message"}}</h1>
```

## Browser Language Detection

The application parses the browser's `Accept-Language` header to determine the user's preferred language. The format is:

```
Accept-Language: en-US,en;q=0.9,de;q=0.8
```

The system:
1. Splits languages by comma
2. Normalizes to 2-letter codes (e.g., `en-US` → `en`)
3. Matches against supported languages
4. Falls back to English if no match is found

## Best Practices

1. **Consistent Key Naming**:
   - Use descriptive, hierarchical key names
   - Group related translations together
   - Example: `form.email`, `form.password`, `form.submit`

2. **Avoid Hardcoded Text**:
   - All user-facing text should be translatable
   - Use translation keys even for English text
   - This makes future language additions easier

3. **Placeholder Values**:
   - Use clear placeholder text that describes the expected input
   - Translate placeholder text as well

4. **Cultural Considerations**:
   - Be aware of cultural differences in greetings, formats, etc.
   - Date and time formats may need special handling
   - Number formats vary by locale

## Troubleshooting

### Language Not Switching

1. Check browser cookies - clear the `language` cookie if needed
2. Verify the translation file exists in `i18n/translations/`
3. Check server logs for translation loading errors
4. Ensure the language code in `SupportedLanguages` matches the file name

### Missing Translations

1. Verify the translation key exists in the JSON file
2. Check for typos in the key path
3. Ensure the JSON structure matches the key path
4. The system falls back to displaying the key name if translation is missing

### Build Errors

1. Run `go mod tidy` to ensure all dependencies are installed
2. Verify all translation files are valid JSON
3. Check that the `//go:embed` directive is present in `i18n.go`
4. Ensure translation files are in the correct directory

## Future Enhancements

Potential improvements for the multilingual system:

1. **User Profile Language**:
   - Store language preference in user profile
   - Override cookie-based detection for logged-in users

2. **Right-to-Left (RTL) Support**:
   - Add support for RTL languages (Arabic, Hebrew, etc.)
   - Dynamic CSS for RTL layouts

3. **Date/Time Localization**:
   - Format dates and times according to user's locale
   - Use libraries like `time.Format` with locale-specific layouts

4. **Number Formatting**:
   - Localize number, currency, and percentage formats
   - Use appropriate thousand separators and decimal points

5. **Dynamic Translation Loading**:
   - Load only the required language file
   - Reduce memory footprint for deployments with many languages

6. **Translation Management**:
   - Web-based translation editor for administrators
   - Export/import functionality for translators
   - Version tracking for translations
