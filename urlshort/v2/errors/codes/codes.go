package codes

import "fmt"

type Code int

const (
	Internal Code = iota
	AlreadyExists
	NotFound
	BadRequest
)

var codeToStr = map[Code]string{
	Internal:      "INTERNAL",
	AlreadyExists: "ALREADY_EXISTS",
	NotFound:      "NOT_FOUND",
	BadRequest:    "BAD_REQUEST",
}

func (c Code) String() string {
	return codeToStr[c]
}

func (c Code) GoString() string {
	return fmt.Sprintf("codes.Code(%s)", c)
}
