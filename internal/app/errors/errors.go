package errors

type ShortenerError struct {
	err error
	msg string
}

func New(err error, msg string) error {
	return ShortenerError{err: err, msg: msg}
}

func (pe ShortenerError) Error() string {
	return pe.err.Error()
}
func (pe ShortenerError) Msg() string {
	return pe.msg
}
func (pe ShortenerError) Unwrap() error {
	return pe.err
}
