package code_generator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodeGenerator_GenerateCreatesFile_DefaultExt(t *testing.T) {
	tmp := t.TempDir()

	engine, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"pkg/main.tpl": []byte("package {{.Module}}\n// {{.ProjectName}}\nconst V = {{.Value}}"),
	}, nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	g := NewCodeGeneratorWithEngine(engine) // default FileExt == ".go"

	opts := Options{
		Module:      "github.com/example/mod",
		ProjectName: "demo",
		OutDir:      tmp,
		Vars: map[string]interface{}{
			"Value": 42,
		},
	}

	out, err := g.Generate(context.Background(), opts, "pkg/main.tpl")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	wantPath := filepath.Join(tmp, "pkg", "main.go")
	if out != wantPath {
		t.Fatalf("unexpected output path: got %q want %q", out, wantPath)
	}

	b, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, "package github.com/example/mod") || !strings.Contains(got, "// demo") || !strings.Contains(got, "V = 42") {
		t.Fatalf("output content unexpected: %q", got)
	}
}

func TestCodeGenerator_GenerateCreatesFile_NoExt(t *testing.T) {
	tmp := t.TempDir()

	engine, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"pkg/main.tpl": []byte("package {{.Module}}\n// {{.ProjectName}}"),
	}, nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	g := NewCodeGeneratorWithEngine(engine)
	g.FileExt = "" // 禁止自动追加扩展名

	opts := Options{
		Module:      "m",
		ProjectName: "p",
		OutDir:      tmp,
	}

	out, err := g.Generate(context.Background(), opts, "pkg/main.tpl")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	wantPath := filepath.Join(tmp, "pkg", "main")
	if out != wantPath {
		t.Fatalf("unexpected output path: got %q want %q", out, wantPath)
	}

	b, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	if !strings.Contains(string(b), "package m") || !strings.Contains(string(b), "// p") {
		t.Fatalf("output content unexpected: %q", string(b))
	}
}

func TestCodeGenerator_RemovesSuffixAndCreatesNestedDirs(t *testing.T) {
	tmp := t.TempDir()

	engine, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"a/b/c/file.tmpl": []byte("name: {{.ProjectName}}\n"),
	}, nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	g := NewCodeGeneratorWithEngine(engine)
	g.FileExt = "" // 保持与模板名移除后缀一致（不追加扩展名）

	opts := Options{
		ProjectName: "nested",
		OutDir:      tmp,
	}

	out, err := g.Generate(context.Background(), opts, "a/b/c/file.tmpl")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	wantPath := filepath.Join(tmp, "a", "b", "c", "file")
	if out != wantPath {
		t.Fatalf("unexpected output path: got %q want %q", out, wantPath)
	}

	b, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	if strings.TrimSpace(string(b)) != "name: nested" {
		t.Fatalf("unexpected file content: %q", string(b))
	}
}

func TestCodeGenerator_NoEngineReturnsError(t *testing.T) {
	g := NewCodeGeneratorWithEngine(nil)
	opts := Options{OutDir: t.TempDir()}

	_, err := g.Generate(context.Background(), opts, "doesnotmatter.tpl")
	if err == nil {
		t.Fatalf("expected error when engine is nil, got nil")
	}
	if !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("expected os.ErrInvalid, got: %v", err)
	}
}

func TestCodeGenerator_OverwriteExistingFile(t *testing.T) {
	tmp := t.TempDir()

	engine, err := NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"dup.tpl": []byte("value: {{.Value}}"),
	}, nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// create existing file with different content
	outPath := filepath.Join(tmp, "dup")
	if err = os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err = os.WriteFile(outPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write initial file failed: %v", err)
	}

	g := NewCodeGeneratorWithEngine(engine)
	g.FileExt = "" // 使输出为 "dup"（无扩展名）

	opts := Options{
		OutDir: tmp,
		Vars: map[string]interface{}{
			"Value": "new",
		},
	}

	out, err := g.Generate(context.Background(), opts, "dup.tpl")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if out != outPath {
		t.Fatalf("unexpected output path: got %q want %q", out, outPath)
	}

	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file failed: %v", err)
	}
	if strings.TrimSpace(string(b)) != "value: new" {
		t.Fatalf("file was not overwritten as expected, got: %q", string(b))
	}
}
