package gonf

import (
	"bytes"
	"log/slog"
	"strconv"
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
	data     string
}

func (t *TemplateValue) Bytes() []byte {
	if t.contents == nil {
		tpl := templates.New("")
		if _, err := tpl.Parse(t.data); err != nil {
			slog.Error("template", "step", "Parse", "data", t.data, "error", err)
			return nil
		}
		var buff bytes.Buffer
		if err := tpl.Execute(&buff, nil /* no data needed */); err != nil {
			slog.Error("template", "step", "Execute", "data", t.data, "error", err)
			return nil
		}
		t.contents = buff.Bytes()
	}
	return t.contents
}
func (t TemplateValue) Int() (int, error) {
	return strconv.Atoi(t.String())
}
func (t TemplateValue) String() string {
	return string(t.Bytes()[:])
}
