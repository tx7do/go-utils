package code_generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// CodeGenerator 使用 TemplateEngine 渲染并将结果写入磁盘
type CodeGenerator struct {
	Engine  TemplateEngine
	FileExt string
}

// NewCodeGeneratorWithEngine 使用指定的引擎创建生成器
func NewCodeGeneratorWithEngine(engine TemplateEngine) *CodeGenerator {
	g := &CodeGenerator{
		Engine:  engine,
		FileExt: ".go",
	}
	return g
}

// Generate 渲染 tplName 并写入 opts.OutDir 下。
// 规则：如果 tplName 以 .tpl 或 .tmpl 结尾，会在输出文件名中去掉该后缀。
func (g *CodeGenerator) Generate(_ context.Context, opts Options, tplName string) (outputPath string, err error) {
	if g.Engine == nil {
		return "", os.ErrInvalid
	}

	// 合并数据：以 opts.Vars 为基础，注入常用字段
	data := map[string]any{}
	if opts.Vars != nil {
		for k, v := range opts.Vars {
			data[k] = v
		}
	}
	// 常用上下文
	data["Module"] = opts.Module
	data["ProjectName"] = opts.ProjectName
	data["Project"] = opts.ProjectName
	data["OutDir"] = opts.OutDir

	// 渲染
	outBytes, err := g.Engine.Render(tplName, data)
	if err != nil {
		return "", err
	}

	// 计算默认输出名称（保持相对目录并去掉模板后缀）
	defaultOutName := tplName
	if strings.HasSuffix(defaultOutName, ".tpl") {
		defaultOutName = strings.TrimSuffix(defaultOutName, ".tpl")
	} else if strings.HasSuffix(defaultOutName, ".tmpl") {
		defaultOutName = strings.TrimSuffix(defaultOutName, ".tmpl")
	}
	defaultOutName = filepath.FromSlash(defaultOutName)

	// 如果用户指定 OutputName，优先处理（规范化、禁止绝对路径）
	finalRel := defaultOutName
	if opts.OutputName != "" {
		user := filepath.FromSlash(opts.OutputName)
		// 如果是绝对路径，去掉根，使之相对（避免写到外部）
		if filepath.IsAbs(user) {
			vol := filepath.VolumeName(user)
			user = strings.TrimPrefix(user, vol)
			user = strings.TrimLeft(user, string(os.PathSeparator))
		}
		if filepath.Dir(user) == "." || filepath.Dir(user) == "" {
			// 仅文件名：保留模板的目录结构，替换基础名
			baseDir := filepath.Dir(defaultOutName)
			finalRel = filepath.Join(baseDir, user)
		} else {
			// 含目录：直接使用用户提供的相对路径
			finalRel = user
		}
	}

	// 清理并防止目录向上穿越
	finalRel = filepath.Clean(finalRel)
	if finalRel == ".." || strings.HasPrefix(finalRel, ".."+string(os.PathSeparator)) {
		return "", os.ErrInvalid
	}

	// 若 finalRel 没有扩展名且 g.FileExt 非空，则追加 FileExt（确保以 '.' 开头）
	if filepath.Ext(finalRel) == "" && g.FileExt != "" {
		fe := g.FileExt
		if !strings.HasPrefix(fe, ".") {
			fe = "." + fe
		}
		finalRel = finalRel + fe
	}

	outPath := filepath.Join(opts.OutDir, finalRel)
	if err = os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", err
	}

	// 原子写入：先写临时文件再重命名
	dir := filepath.Dir(outPath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return "", err
	}
	tmpName := tmpFile.Name()

	_, err = tmpFile.Write(outBytes)
	if errClose := tmpFile.Close(); err == nil {
		err = errClose
	}
	if err != nil {
		_ = os.Remove(tmpName)
		return "", err
	}

	if err = os.Rename(tmpName, outPath); err != nil {
		_ = os.Remove(tmpName)
		return "", err
	}

	// 确保目标文件权限
	_ = os.Chmod(outPath, 0o644)

	return outPath, nil
}
