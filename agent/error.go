package agent

// Err tags errors encountered by the agent
type Err struct {
	E  error
	Op string
}

func (e Err) Error() string { return e.E.Error() }
