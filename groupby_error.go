package main

type GroupbyError struct {
	Message string
}

func groupbyError(msg string) GroupbyError {
	return GroupbyError{
		Message: msg,
	}
}

func (e GroupbyError) Error() string {
	return e.Message
}
