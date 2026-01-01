package main

import (
"embed"
"io/fs"
"testing"

"github.com/swissmakers/wireguard-manager/router"
)

//go:embed templates/*
var embeddedTemplates embed.FS

// TestRouterInitialization verifies that the router initializes without panic
// and that all templates, including the newly added security templates, load correctly
func TestRouterInitialization(t *testing.T) {
// Strip the "templates/" prefix from the embedded templates directory.
tmplDir, err := fs.Sub(embeddedTemplates, "templates")
if err != nil {
t.Fatalf("Error processing templates: %v", err)
}

extraData := map[string]interface{}{
"appVersion":    "test",
"gitCommit":     "test",
"basePath":      "/",
"loginDisabled": false,
}

var secret [64]byte
copy(secret[:], []byte("test-secret-key-for-template-testing-only-123456789"))

// This should not panic if all templates are loaded correctly
app := router.New(tmplDir, extraData, secret)

if app == nil {
t.Fatal("Router initialization returned nil")
}

if app.Renderer == nil {
t.Fatal("Template renderer is not configured")
}

t.Log("✅ Templates loaded successfully!")
t.Log("✅ Router initialized without panic!")
t.Logf("✅ Total routes registered: %d", len(app.Routes()))
}
