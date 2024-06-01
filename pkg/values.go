package gonf

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

type Value interface {
	Bytes() []byte
	Int() (int, error)
	String() string
}

func interfaceToValue(v any) Value {
	switch vv := v.(type) {
	case []byte:
		return &BytesValue{vv}
	case int:
		return &IntValue{vv}
	case string:
		return &StringValue{vv}
	case *VariablePromise:
		return vv
	default:
		slog.Error("interfaceToValue", "value", v, "error", "Not Implemented")
		panic(fmt.Sprintf("interfaceToValue cannot take type %T as argument. Value was %#v.", v, v))
	}
}

func interfaceToTemplateValue(v any) Value {
	if vv, ok := v.([]byte); ok {
		return &TemplateValue{data: string(vv)}
	}
	if vv, ok := v.(int); ok {
		return &IntValue{vv}
	}
	if vv, ok := v.(string); ok {
		return &TemplateValue{data: vv}
	}
	if vv, ok := v.(*VariablePromise); ok {
		return vv
	}
	slog.Error("interfaceToTemplateValue", "value", v, "error", "Not Implemented")
	panic(fmt.Sprintf("interfaceToTemplateValue cannot take type %T as argument. Value was %#v.", v, v))
}

type BytesValue struct {
	value []byte
}

func (b BytesValue) Bytes() []byte {
	return b.value
}
func (b BytesValue) Int() (int, error) {
	return strconv.Atoi(string(b.value))
}
func (b BytesValue) String() string {
	return string(b.value[:])
}

type IntValue struct {
	value int
}

func (i IntValue) Bytes() []byte {
	return []byte(fmt.Sprint(i.value))
}
func (i IntValue) Int() (int, error) {
	return i.value, nil
}
func (i IntValue) String() string {
	return fmt.Sprint(i.value)
}

type StringsListValue struct {
	value []string
}

func (s *StringsListValue) Append(v ...string) {
	s.value = append(s.value, v...)
}
func (s StringsListValue) Bytes() []byte {
	return []byte(s.String())
}
func (s StringsListValue) Int() (int, error) {
	return len(s.value), nil
}
func (s StringsListValue) String() string {
	return strings.Join(s.value, "\n")
}

type StringValue struct {
	value string
}

func (s StringValue) Bytes() []byte {
	return []byte(s.value)
}
func (s StringValue) Int() (int, error) {
	return strconv.Atoi(s.value)
}
func (s StringValue) String() string {
	return s.value
}

// TODO maps

// TODO what else?
