package gonf

import (
	"fmt"
	"log/slog"
)

var variables map[string]*VariablePromise

func init() {
	variables = make(map[string]*VariablePromise)
}

func AppendVariable(name string, values ...string) *VariablePromise {
	if v, ok := variables[name]; ok {
		if l, ok := v.value.(*StringsListValue); ok {
			l.Append(values...)
		}
		return v
	}
	v := &VariablePromise{
		isDefault: false,
		name:      name,
		value:     &StringsListValue{values},
	}
	variables[name] = v
	return v
}

func Default(name string, value string) *VariablePromise {
	if v, ok := variables[name]; ok {
		if !v.isDefault {
			slog.Debug("default would overwrite a variable, ignoring", "name", name, "old_value", v.value, "new_value", value)
			return nil
		}
		slog.Error("default is being overwritten", "name", name, "old_value", v.value, "new_value", value)
	}
	v := &VariablePromise{
		isDefault: true,
		name:      name,
		value:     interfaceToTemplateValue(value),
	}
	variables[name] = v
	return v
}

func Variable(name string, value string) *VariablePromise {
	if v, ok := variables[name]; ok && !v.isDefault {
		slog.Error("variable is being overwritten", "name", name, "old_value", v, "new_value", value)
	}
	v := &VariablePromise{
		isDefault: false,
		name:      name,
		value:     interfaceToTemplateValue(value),
	}
	variables[name] = v
	return v
}

type VariablePromise struct {
	isDefault bool
	name      string
	value     Value
}

func (s VariablePromise) Bytes() []byte {
	return s.value.Bytes()
}
func (s VariablePromise) Int() (int, error) {
	return s.value.Int()
}
func (s VariablePromise) String() string {
	return s.value.String()
}

func getVariable(name string) string {
	if v, ok := variables[name]; ok {
		return v.value.String()
	} else {
		slog.Error("undefined variable or default", "name", name)
		panic(fmt.Sprintf("undefined variable or default %s", name))
	}
}
