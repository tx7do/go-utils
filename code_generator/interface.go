package code_generator

import (
	"context"
	"text/template"
)

// Options 生成选项：包含模块名、项目名、输出目录和变量等
type Options struct {
	// Module 	模块名，例如 github.com/user/project
	Module string
	// ProjectName 项目名称，例如 project
	ProjectName string

	// OutDir 输出目录
	OutDir string
	// OutputName 输出文件名（可选），如果为空则使用模板名去掉后缀作为文件名
	OutputName string

	// Vars 额外变量映射，可在模板中使用
	Vars map[string]interface{}
}

// Generator 通用生成器：渲染指定模板并写入输出
type Generator interface {
	// Generate 根据选项渲染指定模板并写入输出目录
	Generate(ctx context.Context, opts Options, tplName string) (outputPath string, err error)
}

// TemplateEngine 模板引擎接口：加载/渲染/列出模板
type TemplateEngine interface {
	// Render 渲染指定模板并返回结果
	Render(tplName string, data any) ([]byte, error)

	// ListTemplates 列出可用的模板名称
	ListTemplates() []string

	// InstallFuncMap 安装自定义函数映射
	InstallFuncMap(funcs template.FuncMap)
}
