package gonf

import (
	"bytes"
	"log/slog"
	"text/template"
)

// ----- Globals ---------------------------------------------------------------
var templates *template.Template

// ----- Init ------------------------------------------------------------------
func init() {
	templates = template.New("")
	templates.Option("missingkey=error")
	templates.Funcs(builtinTemplateFunctions)
}

// ----- Public ----------------------------------------------------------------
type TemplateValue struct {
	contents []byte
	name     string
}

func (t *TemplateValue) Bytes() []byte {
	var buff bytes.Buffer
	if err := templates.ExecuteTemplate(&buff, t.name, nil /* no data needed */); err != nil {
		slog.Error("template", "step", "ExecuteTemplate", "name", t.name, "error", err)
		return nil
	}
	return buff.Bytes()
}

func (t TemplateValue) String() string {
	return string(t.Bytes()[:])
}

func Template(name string, contents []byte) *TemplateValue {
	tpl := templates.New(name)
	if _, err := tpl.Parse(string(contents)); err != nil {
		slog.Error("template", "step", "Parse", "name", name, "error", err)
		return nil
	}
	return &TemplateValue{contents, name}
}
