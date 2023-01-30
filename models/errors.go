package models

type FormError struct {
	Bool bool
	Message string
	Field string
}

type ValidateError struct{
    Message string
}

func (e *ValidateError) Error() string{
    return e.Message
}

type PathError struct{
    Message string
}

func (e *PathError) Error() string{
    return e.Message
}

type AuthError struct{
    Message string
}

func (e *AuthError) Error() string{
    return e.Message
}
