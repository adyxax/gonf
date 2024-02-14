package gonf

import (
	"fmt"
	"log/slog"
)

type Value interface {
	Bytes() []byte
	String() string
}

func interfaceToValue(v any) Value {
	if vv, ok := v.([]byte); ok {
		return &BytesValue{vv}
	}
	if vv, ok := v.(int); ok {
		return &IntValue{vv}
	}
	if vv, ok := v.(string); ok {
		return &StringValue{vv}
	}
	if vv, ok := v.(*VariablePromise); ok {
		return vv
	}
	slog.Error("interfaceToValue", "value", v, "error", "Not Implemented")
	panic(fmt.Sprintf("interfaceToValue cannot take type %T as argument. Value was %#v.", v, v))
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

// ----- BytesValue ------------------------------------------------------------
type BytesValue struct {
	value []byte
}

func (b BytesValue) Bytes() []byte {
	return b.value
}
func (b BytesValue) String() string {
	return string(b.value[:])
}

// ----- IntValue --------------------------------------------------------------
type IntValue struct {
	value int
}

func (i IntValue) Bytes() []byte {
	return []byte(string(i.value))
}
func (i IntValue) Int() int {
	return i.value
}
func (i IntValue) String() string {
	return string(i.value)
}

// ----- StringValue -----------------------------------------------------------
type StringValue struct {
	value string
}

func (s StringValue) Bytes() []byte {
	return []byte(s.value)
}
func (s StringValue) String() string {
	return s.value
}

// TODO lists

// TODO maps

// TODO what else?
