# code_generator

English | [简体中文](README.zh-CN.md)

`code_generator` provides a general-purpose code/text generator built on Go's `text/template`: it renders templates from an in-memory set or a directory on disk and writes the results safely into an output directory under the template's relative path. The generator and the template engine are decoupled; both are provided as interfaces/structs and can be combined freely.

## Features

- **Decoupled engine and generator**: `TemplateEngine` takes care of "load/render/list templates", while `CodeGenerator` handles "render + write to disk"; each has independent implementations;
- **Two engine implementations**: `EmbeddedTemplateEngine` (built from a `map[string][]byte`, a natural fit for `go:embed`) and `FileTemplateEngine` (recursively loads `.tpl` / `.tmpl` files from a directory on disk);
- **Convention-over-configuration output paths**: automatically strips the `.tpl` / `.tmpl` template suffix, preserves the directory structure embedded in the template name, and appends an extension when needed (`.go` by default);
- **Safe writing to disk**: rejects `..` directory traversal, refuses absolute paths that would escape the output directory, writes atomically via a temporary file plus rename, with `0755` directories and `0644` files;
- **Custom template functions**: inject a `template.FuncMap` when constructing an engine, or override function implementations after the fact with `InstallFuncMap`.

## Installation

```bash
go get github.com/tx7do/go-utils/code_generator
```

```go
import "github.com/tx7do/go-utils/code_generator"
```

The module depends only on the Go standard library (`text/template`, etc.) — no third-party dependencies.

## Quick Start

### Embedded templates + generator

```go
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/tx7do/go-utils/code_generator"
)

func main() {
	// 1. Build an engine from in-memory templates (keys are relative-path style,
	//    '/'-separated; a natural fit with go:embed)
	engine, err := code_generator.NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"pkg/main.tpl": []byte("package {{.ProjectName}}\n// module: {{.Module}}\nconst V = {{.Value}}"),
	}, template.FuncMap{
		"toupper": strings.ToUpper, // optional: inject custom functions before rendering
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create the generator (FileExt defaults to ".go")
	gen := code_generator.NewCodeGeneratorWithEngine(engine)

	// 3. Render "pkg/main.tpl" and write it into the output directory
	outPath, err := gen.Generate(context.Background(), code_generator.Options{
		Module:      "github.com/example/mod",
		ProjectName: "demo",
		OutDir:      "gen",
		Vars:        map[string]interface{}{"Value": 42},
	}, "pkg/main.tpl")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(outPath) // gen/pkg/main.go
}
```

Besides `Vars`, four common keys are always injected into the template data: `Module`, `ProjectName`, `Project` (same value as `ProjectName`), and `OutDir`.

### A directory of file templates

```go
engine, err := code_generator.NewFileTemplateEngine("templates") // recursively loads *.tpl, *.tmpl
if err != nil {
	log.Fatal(err)
}

fmt.Println(engine.ListTemplates()) // e.g. [hello.tpl sub/service.tmpl]

out, err := engine.Render("hello.tpl", map[string]string{"Name": "world"})
// out == []byte("Hello world") (when the template contains "Hello {{.Name}}")

// InstallFuncMap re-parses the template trees, overriding same-named function
// implementations:
engine.InstallFuncMap(template.FuncMap{"up": strings.ToUpper})
```

## Output Path Rules

`CodeGenerator.Generate(ctx, opts, tplName)` determines the output file in the following order (see `code_generator.go`; every rule below is backed by tests):

1. **Default output name = template name with the `.tpl` / `.tmpl` suffix removed**: `pkg/main.tpl` → `pkg/main`, `a/b/c/file.tmpl` → `a/b/c/file`. Any directory structure in the template name is preserved;
2. **`opts.OutputName` overrides**:
   - A bare file name (no `/`): keeps the directory from the template name and replaces only the base name (`pkg/main.tpl` + `OutputName: "user"` → `pkg/user`);
   - Contains a directory: used as the relative path as-is (`OutputName: "sub/user"` → `sub/user`);
   - An absolute path: the volume/root is stripped and it is treated as a relative path (any escape from `OutDir` is then caught by the traversal check below);
3. **Directory-traversal guard**: after `Clean`, if the result is `..` or starts with `../`, `os.ErrInvalid` is returned;
4. **Extension completion**: when the result has no extension and `g.FileExt != ""`, `g.FileExt` is appended (a leading `.` is added if missing); the constructor defaults it to `.go` — set it to `""` to disable completion entirely. An existing extension (e.g. a template that already ends in `.json` after suffix removal) is never overwritten;
5. **Write into `opts.OutDir`**: missing directories are created automatically (`0755`); the write is atomic via a temporary file followed by `Rename`; the final file is set to `0644`. An existing file is **overwritten outright**.

## Template Name Matching

Both engines resolve `Render(tplName, data)` names through the same fallback chain (tests: `embed_template_engine_test.go`, `file_template_engine_test.go`):

1. Exact match against a registered key (e.g. `sub/main.tpl`);
2. Match after stripping a leading `./`;
3. Suffix match: any key ending in `/<tplName>` matches (e.g. with `sub/main.tpl` registered, `main.tpl` also resolves).

When nothing matches, `EmbeddedTemplateEngine` returns `template not found: <tplName>` and `FileTemplateEngine` returns `os.ErrNotExist`.

## API Reference

- `type Options struct` — generation options: `Module` (module name), `ProjectName` (project name), `OutDir` (output directory), `OutputName` (optional override of the output file name), `Vars map[string]interface{}` (extra template variables);
- `type Generator interface` — `Generate(ctx context.Context, opts Options, tplName string) (outputPath string, err error)`;
- `type TemplateEngine interface` — `Render(tplName string, data any) ([]byte, error)`, `ListTemplates() []string`, `InstallFuncMap(funcs template.FuncMap)`;
- `type CodeGenerator struct` — fields `Engine TemplateEngine`, `FileExt string`;
- `func NewCodeGeneratorWithEngine(engine TemplateEngine) *CodeGenerator` — `FileExt` defaults to `".go"`;
- `func (*CodeGenerator) Generate(ctx context.Context, opts Options, tplName string) (string, error)` — renders and writes to disk; returns `os.ErrInvalid` when `Engine` is nil (`ctx` is currently unused);
- `type EmbeddedTemplateEngine struct`;
- `func NewEmbeddedTemplateEngine(srcs map[string][]byte) (*EmbeddedTemplateEngine, error)` — equivalent to `NewEmbeddedTemplateEngineFromMap(srcs, nil)`;
- `func NewEmbeddedTemplateEngineFromMap(srcs map[string][]byte, funcs template.FuncMap) (*EmbeddedTemplateEngine, error)` — parses templates one by one (empty entries are skipped); `funcs` are injected before parsing;
- `type FileTemplateEngine struct`;
- `func NewFileTemplateEngine(root string) (*FileTemplateEngine, error)` — recursively loads every `.tpl` / `.tmpl` under `root` (an empty `root` is treated as `"."`); keys are `/`-separated relative paths.

All three `TemplateEngine` methods (`Render` / `ListTemplates` / `InstallFuncMap`) are implemented and exported on both implementations.

## Notes

- **Built-in keys win over `Vars`**: the render data starts as a copy of `Vars`, then `Module` / `ProjectName` / `Project` / `OutDir` are written unconditionally — a `Vars` key with the same name is overwritten by these four built-ins;
- **`ctx` has no effect**: the context argument of `CodeGenerator.Generate` is ignored (the parameter is `_`), so a write cannot be cancelled through it;
- **File templates are loaded once at construction**: changes on disk after `NewFileTemplateEngine` are not seen; rebuild the engine to pick them up;
- **Template functions must exist at parse time**: Go's `text/template` validates function names during parsing, so a template referencing an unregistered function makes **engine construction fail** — `EmbeddedTemplateEngine` requires providing them via the `funcs` argument of `NewEmbeddedTemplateEngineFromMap`; `FileTemplateEngine` has no construction parameter for a `FuncMap`, so it cannot load templates that use custom functions;
- **`InstallFuncMap` works by re-parsing the template trees**: for templates that parsed successfully it installs/overrides same-named function implementations (verified: the implementation of `toupper` can be swapped wholesale); templates whose re-parse fails are left untouched (silently skipped). It cannot rescue an engine that failed to build over a missing function — that construction already errored;
- **Suffix-match ordering is not guaranteed**: name matching iterates a map, so when several keys end with `/<tplName>` it is nondeterministic which one wins — pass the full key when in doubt;
- **Overwrites**: `Generate` never confirms before replacing an existing output file; it is atomically overwritten;
- **`FileExt` is appended only when there is no extension**: if the template name still carries an extension after `.tpl` / `.tmpl` removal (e.g. `x.yaml.tpl` → `x.yaml`), `FileExt` is not appended.
