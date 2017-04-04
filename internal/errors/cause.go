package errors

// Causer is an interface for error wrappers to expose their wrapped
// error.
type Causer interface {
	error
	Cause() error
}

// Cause returns the innermost error cause: it keeps trying to unwrap
// err as a Causer, until it can't anymore, finally returning the
// error.
func Cause(err error) error {
	for c, ok := err.(Causer); ok; c, ok = err.(Causer) {
		err = c.Cause()
	}
	return err
}
