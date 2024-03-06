package gonf

type Promise interface {
	IfRepaired(...Promise) Promise
	Promise() Promise
	Resolve()
}

//type Operation int
//
//const (
//	AND = iota
//	OR
//	NOT
//)
//
//func (o Operation) String() string {
//	switch o {
//	case AND:
//		return "and"
//	case OR:
//		return "or"
//	case NOT:
//		return "not"
//	}
//	panic("unknown")
//}

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
