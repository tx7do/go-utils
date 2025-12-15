package code_generator

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// FileTemplateEngine 从磁盘加载并缓存模板
type FileTemplateEngine struct {
	root      string
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewFileTemplateEngine 创建并预加载模板目录（支持 .tpl/.tmpl 后缀）
func NewFileTemplateEngine(root string) (*FileTemplateEngine, error) {
	e := &FileTemplateEngine{
		root:      root,
		templates: make(map[string]*template.Template),
	}
	if root == "" {
		root = "."
	}
	err := e.loadAll()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// loadAll 加载 root 目录下的所有模板文件
func (e *FileTemplateEngine) loadAll() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return filepath.WalkDir(e.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !(strings.HasSuffix(name, ".tpl") || strings.HasSuffix(name, ".tmpl")) {
			return nil
		}
		rel, err := filepath.Rel(e.root, path)
		if err != nil {
			return err
		}
		// use slash-separated template key
		key := filepath.ToSlash(rel)
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return err
		}
		e.templates[key] = tmpl
		return nil
	})
}

// Render 渲染指定模板（通过相对于 root 的路径名，如 "service/main.tpl"）
func (e *FileTemplateEngine) Render(tplName string, data any) ([]byte, error) {
	e.mu.RLock()
	tmpl, ok := e.templates[tplName]
	e.mu.RUnlock()

	// 如果未找到，尝试尝试带/不带前缀的匹配（简单容错）
	if !ok {
		// 尝试去掉前导"./"
		alt := strings.TrimPrefix(tplName, "./")
		e.mu.RLock()
		tmpl, ok = e.templates[alt]
		e.mu.RUnlock()
	}
	if !ok {
		// 最后尝试查找同名文件在子目录下（简单尝试）
		e.mu.RLock()
		for k, t := range e.templates {
			if strings.HasSuffix(k, "/"+tplName) || k == tplName {
				tmpl = t
				ok = true
				break
			}
		}
		e.mu.RUnlock()
	}

	if !ok {
		return nil, os.ErrNotExist
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ListTemplates 列出可用模板名（相对于 root 的路径，使用 '/' 分隔）
func (e *FileTemplateEngine) ListTemplates() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.templates))
	for k := range e.templates {
		out = append(out, k)
	}
	return out
}

func (e *FileTemplateEngine) InstallFuncMap(funcs template.FuncMap) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for name, tmpl := range e.templates {
		t := tmpl.Funcs(funcs)
		parsed, err := t.Parse(tmpl.Tree.Root.String())
		if err == nil {
			e.templates[name] = parsed
		}
	}
}
