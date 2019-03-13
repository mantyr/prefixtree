package prefix

const (
	Static Type = iota
	Root
	Param
	CatchAll
)

// Type эти тип ноды
type Type uint8
