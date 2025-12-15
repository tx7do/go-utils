package code_generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func writeFile(t *testing.T, base, rel, content string) {
	t.Helper()
	path := filepath.Join(base, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writefile: %v", err)
	}
}

func TestFileTemplateEngine_LoadListRenderAndInstallFuncMap(t *testing.T) {
	td := t.TempDir()

	// create templates
	writeFile(t, td, "hello.tpl", "Hello {{.Name}}")
	writeFile(t, td, "sub/service.tmpl", "Service: {{.Svc}}")
	writeFile(t, td, "func.tpl", "Upper: {{up .Val}}")

	engine, err := NewFileTemplateEngine(td)
	if err != nil {
		t.Fatalf("NewFileTemplateEngine error: %v", err)
	}

	// ListTemplates contains expected keys
	templates := engine.ListTemplates()
	if !contains(templates, "hello.tpl") {
		t.Fatalf("missing hello.tpl in templates: %v", templates)
	}
	if !contains(templates, "sub/service.tmpl") {
		t.Fatalf("missing sub/service.tmpl in templates: %v", templates)
	}
	if !contains(templates, "func.tpl") {
		t.Fatalf("missing func.tpl in templates: %v", templates)
	}

	// Render simple template
	out, err := engine.Render("hello.tpl", map[string]string{"Name": "World"})
	if err != nil {
		t.Fatalf("Render hello.tpl error: %v", err)
	}
	if string(out) != "Hello World" {
		t.Fatalf("unexpected hello.tpl output: %q", string(out))
	}

	// Render with ./ prefix (fallback)
	out, err = engine.Render("./hello.tpl", map[string]string{"Name": "X"})
	if err != nil {
		t.Fatalf("Render ./hello.tpl error: %v", err)
	}
	if string(out) != "Hello X" {
		t.Fatalf("unexpected ./hello.tpl output: %q", string(out))
	}

	// Render by basename to match suffix-in-dir fallback (should find sub/service.tmpl)
	out, err = engine.Render("service.tmpl", map[string]string{"Svc": "API"})
	if err != nil {
		t.Fatalf("Render service.tmpl error: %v", err)
	}
	if string(out) != "Service: API" {
		t.Fatalf("unexpected service.tmpl output: %q", string(out))
	}

	// Missing template returns os.ErrNotExist
	_, err = engine.Render("nope.tpl", nil)
	if err == nil {
		t.Fatalf("expected error for missing template")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("expected os.ErrNotExist, got: %v", err)
	}

	// func.tpl should fail before installing func map
	_, err = engine.Render("func.tpl", map[string]string{"Val": "hello"})
	if err == nil {
		t.Fatalf("expected error rendering func.tpl without func map")
	}

	// install func map and render
	engine.InstallFuncMap(template.FuncMap{
		"up": strings.ToUpper,
	})
	out, err = engine.Render("func.tpl", map[string]string{"Val": "hello"})
	if err != nil {
		t.Fatalf("Render func.tpl after InstallFuncMap error: %v", err)
	}
	if string(out) != "Upper: HELLO" {
		t.Fatalf("unexpected func.tpl output after InstallFuncMap: %q", string(out))
	}
}
