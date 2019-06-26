package httpt

type TransportErr struct {
	E  error
	Op string
}

func (e TransportErr) Error() string { return e.E.Error() }

// var (
// 	ErrInvalidRequest = errors.New("invalid request parameters")
// 	ErrRequestFailed  = errors.New("request failed")
// )
