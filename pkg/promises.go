package gonf

type Promise interface {
	IfRepaired(...Promise) Promise
	Resolve()
	Status() Status
}

type Status int

const (
	PROMISED = iota
	BROKEN
	KEPT
	REPAIRED
)

func (s Status) String() string {
	switch s {
	case PROMISED:
		return "promised"
	case BROKEN:
		return "broken"
	case KEPT:
		return "kept"
	case REPAIRED:
		return "repaired"
	}
	panic("unknown")
}
