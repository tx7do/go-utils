package code_generator

import (
	"strings"
	"testing"
	"text/template"
)

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestEmbeddedTemplateEngine_RenderAndList(t *testing.T) {
	tmplMain := []byte("hello {{.Name}}")
	tmplSub := []byte("sub {{.Name}}")

	engine, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"main.tpl":     tmplMain,
		"sub/main.tpl": tmplSub,
	}, nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	t.Run("Render main.tpl", func(t *testing.T) {
		out, err := engine.Render("main.tpl", map[string]string{"Name": "world"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := strings.TrimSpace(string(out)); got != "hello world" {
			t.Fatalf("render result = %q, want %q", got, "hello world")
		}
	})

	t.Run("Render with ./ prefix", func(t *testing.T) {
		out, err := engine.Render("./main.tpl", map[string]string{"Name": "dot"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := strings.TrimSpace(string(out)); got != "hello dot" {
			t.Fatalf("render result = %q, want %q", got, "hello dot")
		}
	})

	t.Run("Suffix match finds sub/main.tpl", func(t *testing.T) {
		// 只包含 sub/main.tpl 的 engine 测试后缀匹配
		e2, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
			"sub/main.tpl": tmplSub,
		}, nil)
		if err != nil {
			t.Fatalf("failed to create engine: %v", err)
		}
		out, err := e2.Render("main.tpl", map[string]string{"Name": "sub"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := strings.TrimSpace(string(out)); got != "sub sub" {
			t.Fatalf("render result = %q, want %q", got, "sub sub")
		}
	})

	t.Run("Render not found", func(t *testing.T) {
		empty, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{}, nil)
		if err != nil {
			t.Fatalf("failed to create engine: %v", err)
		}
		_, err = empty.Render("no.tpl", nil)
		if err == nil {
			t.Fatalf("expected error for missing template, got nil")
		}
	})

	t.Run("ListTemplates contains keys", func(t *testing.T) {
		list := engine.ListTemplates()
		if !contains(list, "main.tpl") || !contains(list, "sub/main.tpl") {
			t.Fatalf("ListTemplates missing expected keys: %v", list)
		}
	})
}

func TestEmbeddedTemplateEngine_FuncMapInjection(t *testing.T) {
	// 模板使用自定义函数 toupper
	engine, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"f.tpl": []byte("UPPER: {{ toupper .Name }}"),
	}, template.FuncMap{
		"toupper": strings.ToUpper,
	})
	if err != nil {
		t.Fatalf("failed to create engine with funcs: %v", err)
	}

	out, err := engine.Render("f.tpl", map[string]string{"Name": "mix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "UPPER: MIX" {
		t.Fatalf("func map not applied, got: %q", got)
	}
}
