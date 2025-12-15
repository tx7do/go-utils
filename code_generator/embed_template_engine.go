package code_generator

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"text/template"
)

// EmbeddedTemplateEngine 从给定的 name->[]byte 映射构建模板集合并提供渲染。
type EmbeddedTemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
}

func NewEmbeddedTemplateEngine(srcs map[string][]byte) (*EmbeddedTemplateEngine, error) {
	return NewEmbeddedTemplateEngineFromMap(srcs, nil)
}

// NewEmbeddedTemplateEngineFromMap 使用预先准备好的模板字节映射创建引擎，
// 可选传入 FuncMap，会在解析模板前调用 template.New(name).Funcs(funcs)。
// 键应为相对路径样式（使用 '/' 分隔），例如 `main.tpl` 或 `sub/main.tpl`。
func NewEmbeddedTemplateEngineFromMap(srcs map[string][]byte, funcs template.FuncMap) (*EmbeddedTemplateEngine, error) {
	e := &EmbeddedTemplateEngine{
		templates: make(map[string]*template.Template),
	}

	for name, b := range srcs {
		if len(b) == 0 {
			continue
		}
		t := template.New(name)
		if funcs != nil {
			t = t.Funcs(funcs)
		}
		tmpl, err := t.Parse(string(b))
		if err != nil {
			return nil, err
		}
		e.templates[name] = tmpl
	}

	return e, nil
}

func (e *EmbeddedTemplateEngine) InstallFuncMap(funcs template.FuncMap) {
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

// Render 渲染指定模板名（例如 "main.tpl" 或 "service.tpl"），返回渲染后的字节。
func (e *EmbeddedTemplateEngine) Render(tplName string, data any) ([]byte, error) {
	e.mu.RLock()
	tmpl, ok := e.templates[tplName]
	e.mu.RUnlock()

	// 容错：去掉前缀 ./ 或尝试只用基础名
	if !ok {
		alt := strings.TrimPrefix(tplName, "./")
		e.mu.RLock()
		tmpl, ok = e.templates[alt]
		e.mu.RUnlock()
	}
	if !ok {
		// 尝试匹配结尾相同的键
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
		return nil, errors.New("template not found: " + tplName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ListTemplates 返回可用模板名称列表（相对于映射的键，如 "main.tpl"）。
func (e *EmbeddedTemplateEngine) ListTemplates() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.templates))
	for k := range e.templates {
		out = append(out, k)
	}
	return out
}
