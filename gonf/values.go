package gonf

import "log/slog"

type Value interface {
	Bytes() []byte
	String() string
}

func interfaceToValue(v any) Value {
	if vv, ok := v.(string); ok {
		return &StringValue{vv}
	}
	if vv, ok := v.([]byte); ok {
		return &BytesValue{vv}
	}
	slog.Error("interfaceToTemplateValue", "value", v, "error", "Not Implemented")
	return nil
}

func interfaceToTemplateValue(v any) Value {
	if vv, ok := v.(string); ok {
		return &TemplateValue{data: vv}
	}
	if vv, ok := v.([]byte); ok {
		return &TemplateValue{data: string(vv)}
	}
	slog.Error("interfaceToTemplateValue", "value", v, "error", "Not Implemented")
	return nil
}

// ----- BytesValue -----------------------------------------------------------------
type BytesValue struct {
	value []byte
}

func (b BytesValue) Bytes() []byte {
	return b.value
}

func (b BytesValue) String() string {
	return string(b.value[:])
}

// ----- StringValue ----------------------------------------------------------------
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
