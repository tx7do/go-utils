# code_generator

[English](README.md) | 简体中文

`code_generator` 提供一个基于 Go `text/template` 的通用代码/文本生成器：用内嵌模板或磁盘模板目录渲染模板，并把结果按模板相对路径安全地写入输出目录。生成器与模板引擎解耦，二者均以接口/结构体形式提供，可自由组合。

## 特性

- **引擎与生成器解耦**：`TemplateEngine` 负责"加载/渲染/列出模板"，`CodeGenerator` 负责"渲染 + 写盘"，各自独立实现；
- **两种引擎实现**：`EmbeddedTemplateEngine`（从 `map[string][]byte` 构建，配合 `go:embed` 使用）与 `FileTemplateEngine`（递归加载磁盘目录下的 `.tpl` / `.tmpl` 文件）；
- **约定优于配置的输出路径**：自动剥离 `.tpl` / `.tmpl` 模板后缀、保留模板名中的目录结构、按需追加扩展名（默认 `.go`）；
- **安全写盘**：禁止 `..` 目录穿越、拒绝绝对路径写入外部、先写临时文件再重命名的原子写入、目录 `0755` / 文件 `0644`；
- **自定义模板函数**：构建引擎时可注入 `template.FuncMap`，事后也可用 `InstallFuncMap` 覆盖函数实现。

## 安装

```bash
go get github.com/tx7do/go-utils/code_generator
```

```go
import "github.com/tx7do/go-utils/code_generator"
```

模块仅依赖 Go 标准库（`text/template` 等），无第三方依赖。

## 快速开始

### 内嵌模板 + 生成器

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
	// 1. 从内存模板构建引擎（键为相对路径样式，用 '/' 分隔；典型配合 go:embed 使用）
	engine, err := code_generator.NewEmbeddedTemplateEngineFromMap(map[string][]byte{
		"pkg/main.tpl": []byte("package {{.ProjectName}}\n// module: {{.Module}}\nconst V = {{.Value}}"),
	}, template.FuncMap{
		"toupper": strings.ToUpper, // 可选：渲染前注入自定义函数
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. 创建生成器（默认 FileExt 为 ".go"）
	gen := code_generator.NewCodeGeneratorWithEngine(engine)

	// 3. 渲染 "pkg/main.tpl" 并写入输出目录
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

模板数据除 `Vars` 外，还会自动注入 `Module`、`ProjectName`、`Project`（与 `ProjectName` 同值）、`OutDir` 四个常用键。

### 文件模板目录

```go
engine, err := code_generator.NewFileTemplateEngine("templates") // 递归加载 *.tpl、*.tmpl
if err != nil {
	log.Fatal(err)
}

fmt.Println(engine.ListTemplates()) // 如 [hello.tpl sub/service.tmpl]

out, err := engine.Render("hello.tpl", map[string]string{"Name": "world"})
// out == []byte("Hello world")（模板内容为 "Hello {{.Name}}" 时）

// InstallFuncMap 会重新解析模板树，覆盖同名函数的实现：
engine.InstallFuncMap(template.FuncMap{"up": strings.ToUpper})
```

## 输出路径规则

`CodeGenerator.Generate(ctx, opts, tplName)` 按以下顺序确定输出文件（实现见 `code_generator.go`，各条均有测试佐证）：

1. **默认输出名 = 模板名去掉 `.tpl` / `.tmpl` 后缀**：`pkg/main.tpl` → `pkg/main`，`a/b/c/file.tmpl` → `a/b/c/file`。模板名中的目录结构原样保留；
2. **`opts.OutputName` 覆盖**：
   - 仅文件名（不含 `/`）：保留模板名中的目录，只替换基础名（`pkg/main.tpl` + `OutputName: "user"` → `pkg/user`）；
   - 含目录：整体作为相对路径使用（`OutputName: "sub/user"` → `sub/user`）；
   - 绝对路径：剥掉盘符/根，强制按相对路径处理（不会写到 `OutDir` 之外的位置由后续穿越检查兜底）；
3. **目录穿越防护**：`Clean` 后若结果为 `..` 或以 `../` 开头，返回 `os.ErrInvalid`；
4. **扩展名补齐**：结果无扩展名且 `g.FileExt != ""` 时追加 `g.FileExt`（自动确保以 `.` 开头）；构造时默认为 `.go`，设为 `""` 可完全关闭补齐。已有扩展名（如模板名去掉后缀后自带 `.json`）不会被覆盖；
5. **写入 `opts.OutDir`**：自动创建缺失目录（`0755`）；先写临时文件再 `Rename` 的原子写入；目标文件以 `0644` 权限收尾。同名文件会被**直接覆盖**。

## 模板名匹配规则

两个引擎的 `Render(tplName, data)` 查找顺序一致（测试见 `embed_template_engine_test.go`、`file_template_engine_test.go`）：

1. 精确匹配注册键（如 `sub/main.tpl`）；
2. 去掉前导 `./` 再匹配；
3. 后缀匹配：任意键以 `/<tplName>` 结尾即命中（如注册了 `sub/main.tpl` 时用 `main.tpl` 也能找到）。

均未命中时返回错误：`EmbeddedTemplateEngine` 返回 `template not found: <tplName>`，`FileTemplateEngine` 返回 `os.ErrNotExist`。

## API 参考

- `type Options struct` —— 生成选项：`Module`（模块名）、`ProjectName`（项目名）、`OutDir`（输出目录）、`OutputName`（可选，覆盖输出文件名）、`Vars map[string]interface{}`（额外模板变量）；
- `type Generator interface` —— `Generate(ctx context.Context, opts Options, tplName string) (outputPath string, err error)`；
- `type TemplateEngine interface` —— `Render(tplName string, data any) ([]byte, error)`、`ListTemplates() []string`、`InstallFuncMap(funcs template.FuncMap)`；
- `type CodeGenerator struct` —— 字段 `Engine TemplateEngine`、`FileExt string`；
- `func NewCodeGeneratorWithEngine(engine TemplateEngine) *CodeGenerator` —— `FileExt` 默认置为 `".go"`；
- `func (*CodeGenerator) Generate(ctx context.Context, opts Options, tplName string) (string, error)` —— 渲染并写盘；`Engine` 为 nil 时返回 `os.ErrInvalid`（`ctx` 目前未被使用）；
- `type EmbeddedTemplateEngine struct`；
- `func NewEmbeddedTemplateEngine(srcs map[string][]byte) (*EmbeddedTemplateEngine, error)` —— 等价于 `NewEmbeddedTemplateEngineFromMap(srcs, nil)`；
- `func NewEmbeddedTemplateEngineFromMap(srcs map[string][]byte, funcs template.FuncMap) (*EmbeddedTemplateEngine, error)` —— 逐条解析模板（空内容跳过），`funcs` 在解析前注入；
- `type FileTemplateEngine struct`；
- `func NewFileTemplateEngine(root string) (*FileTemplateEngine, error)` —— 递归加载 `root` 下所有 `.tpl` / `.tmpl`（`root` 为空按 `"."` 处理），键为以 `/` 分隔的相对路径。

`TemplateEngine` 的三个方法（`Render` / `ListTemplates` / `InstallFuncMap`）在两个实现上均已导出实现。

## 注意事项

- **内置键优先于 `Vars`**：渲染数据先复制 `Vars`，再无条件写入 `Module` / `ProjectName` / `Project` / `OutDir`，同名 `Vars` 键会被这四个内置键覆盖；
- **`ctx` 未生效**：`CodeGenerator.Generate` 的 context 参数被忽略（形参为 `_`），无法通过它取消写入；
- **文件模板只在构建时加载一次**：`NewFileTemplateEngine` 构造后磁盘上的模板改动不会被感知，需重新构建引擎；
- **模板函数必须在解析时已存在**：Go `text/template` 在解析阶段即校验函数名，模板引用未注册函数会导致**引擎构建失败**——`EmbeddedTemplateEngine` 需通过 `NewEmbeddedTemplateEngineFromMap` 的 `funcs` 参数提供；`FileTemplateEngine` 没有注入 `FuncMap` 的构造参数，因此无法加载引用自定义函数的模板；
- **`InstallFuncMap` 通过重新解析模板树生效**：对已成功解析的模板补装/覆盖同名函数的实现（实测可将 `toupper` 的实现整体替换），重新解析失败的模板保持原样（静默跳过）；它无法拯救构建时因缺函数而失败的引擎——那时构建早已报错；
- **后缀匹配的顺序不保证**：模板名匹配按 map 遍历，存在多个以 `/<tplName>` 结尾的键时命中哪一个是不确定的，建议传完整键名；
- **覆盖写**：`Generate` 对已存在的输出文件不做任何确认，直接原子覆盖；
- **`FileExt` 只在"无扩展名"时追加**：模板名去掉 `.tpl` / `.tmpl` 后若自带扩展名（如 `x.yaml.tpl` → `x.yaml`），不会追加 `FileExt`。
