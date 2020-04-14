package rabbit

import "fmt"

// compile time checks
var (
	_ error = Marshal{}
)

const (
	ErrMarshal int = iota
)

type Marshal struct {
	E   error
	Msg string
}

func (m Marshal) Error() string {
	return m.Msg + m.E.Error()
}

func MarshalError(e error, data []byte) Marshal {
	return Marshal{
		E:   e,
		Msg: fmt.Sprintf("could not unmarshal %s, to tx slice", data),
	}
}
