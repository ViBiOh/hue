package model

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrInvalid occurs when checks fails
	ErrInvalid = errors.New("invalid")

	// ErrNotFound occurs when something is not found
	ErrNotFound = errors.New("not found")

	// ErrMethodNotAllowed occurs when method is not allowed
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrInternalError occurs when shit happens
	ErrInternalError = errors.New("internal error")
)

// TemplateFunc handle a request and returns which template to render with which status and datas
type TemplateFunc = func(*http.Request) (string, int, map[string]interface{}, error)

// Message for render
type Message struct {
	Level   string
	Content string
}

// ParseMessage parses messages from request
func ParseMessage(r *http.Request) Message {
	values := r.URL.Query()

	return Message{
		Level:   values.Get("messageLevel"),
		Content: values.Get("messageContent"),
	}
}

// NewSuccessMessage create a success message
func NewSuccessMessage(content string) Message {
	return Message{
		Level:   "success",
		Content: content,
	}
}

// NewErrorMessage create a error message
func NewErrorMessage(content string) Message {
	return Message{
		Level:   "error",
		Content: content,
	}
}

// ConcatError concat errors to a single string
func ConcatError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	values := make([]string, len(errs))
	for index, err := range errs {
		values[index] = err.Error()
	}

	return errors.New(strings.Join(values, ", "))
}

func wrapError(err, wrapper error) error {
	return fmt.Errorf("%s: %w", err, wrapper)
}

// WrapInvalid wraps given error with invalid err
func WrapInvalid(err error) error {
	return wrapError(err, ErrInvalid)
}

// WrapNotFound wraps given error with not found err
func WrapNotFound(err error) error {
	return wrapError(err, ErrNotFound)
}

// WrapMethodNotAllowed wraps given error with not method not allowed err
func WrapMethodNotAllowed(err error) error {
	return wrapError(err, ErrMethodNotAllowed)
}

// WrapInternal wraps given error with internal err
func WrapInternal(err error) error {
	return wrapError(err, ErrInternalError)
}
