package apperr

type ValidationError struct {
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func NewValidation(msg string) *ValidationError {
	return &ValidationError{Message: msg}
}

func NewValidationWrap(msg string, err error) *ValidationError {
	return &ValidationError{Message: msg, Err: err}
}
