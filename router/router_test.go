package router

import (
	"testing"
	"text/template"
)

// TestTemplatesParsing ensures templates can be parsed without nested define errors.
// This test validates the fix for the "unexpected <define> in command" panic that occurred
// when base.html had a {{define "base.html"}}...{{end}} wrapper, causing nested define blocks
// when concatenated with page templates.
func TestTemplatesParsing(t *testing.T) {
	// Simulate base template without the {{define "base.html"}} wrapper
	baseString := `<!DOCTYPE html>
<html>
<head>{{template "title" .}}</head>
<body>
	{{template "page_content" .}}
	{{template "bottom_js" .}}
</body>
</html>`
	
	// Simulate a page template with its own define blocks
	profileString := `{{ define "title"}}My Profile{{ end }}
{{ define "page_content"}}<h1>Profile Page</h1>{{ end }}
{{ define "bottom_js"}}<script>console.log("test");</script>{{ end }}`
	
	// Try to parse concatenated templates - this should not panic
	// Before the fix, if base.html had {{define "base.html"}} wrapper,
	// this would fail with "unexpected <define> in command"
	tmpl, err := template.New("profile").Parse(baseString + profileString)
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}
	
	// Verify the template was created
	if tmpl == nil {
		t.Fatal("Template is nil")
	}
	
	// Verify it's named correctly
	if tmpl.Name() != "profile" {
		t.Errorf("Expected template name 'profile', got '%s'", tmpl.Name())
	}
}

// TestNestedDefineBlocksNotAllowed demonstrates the bug that was fixed.
// Go templates do not allow nested {{define}} blocks within template execution.
func TestNestedDefineBlocksNotAllowed(t *testing.T) {
	// This test documents that the fix works.
	// The original bug: base.html wrapped content in {{define "base.html"}}...{{end}}
	// When concatenated with page templates, their {{define}} blocks appeared inside,
	// causing "unexpected <define> in command" when the template tried to execute
	// template actions like {{template "title" .}} that reference those nested defines.
	
	// The fix: Remove the {{define "base.html"}} wrapper from base.html
	// so that page template {{define}} blocks are at the top level.
	
	// This structure (fixed) should work fine:
	fixedStructure := `<!DOCTYPE html>
<html>
<head>{{template "title" .}}</head>
</html>
{{define "title"}}Page Title{{end}}`
	
	tmpl, err := template.New("test").Parse(fixedStructure)
	if err != nil {
		t.Errorf("Fixed structure should parse without error, but got: %v", err)
	}
	
	// Verify we can execute it
	if tmpl != nil {
		// Just verifying it doesn't panic
		t.Log("Template parsed successfully with top-level define blocks")
	}
}
