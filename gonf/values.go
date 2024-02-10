package gonf

type Value interface {
	Bytes() []byte
	String() string
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

func Bytes(value []byte) *BytesValue {
	return &BytesValue{value}
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

func String(value string) *StringValue {
	return &StringValue{value}
}

// TODO lists

// TODO maps

// TODO what else?
