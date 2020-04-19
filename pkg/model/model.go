package model

// Message for render
type Message struct {
	Level   string
	Content string
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
