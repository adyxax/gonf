package gonf

type Value interface {
	Equals(Value) bool
	String() string
}

// ----- String variables ------------------------------------------------------
type StringValue struct {
	Value string
}

func (s StringValue) Equals(v Value) bool {
	sv, ok := v.(StringValue)
	return ok && s.Value == sv.Value
}

func (s StringValue) String() string {
	// TODO handle interpolation
	return s.Value
}

// TODO lists

// TODO maps

// TODO what else?
