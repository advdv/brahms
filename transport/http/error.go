package httpt

// TransportErr describes an error during transport functions
type TransportErr struct {
	E  error
	Op string
}

func (e TransportErr) Error() string {
	return e.E.Error()
}
