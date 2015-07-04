package action

type Pattern int

type Action struct {
	Pattern Pattern
	Param   interface{}
}
