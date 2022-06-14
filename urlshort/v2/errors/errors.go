package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/marcuscaisey/gophercises/urlshort/v2/errors/codes"
)

type Error struct {
	msg     string
	code    codes.Code
	wrapped error
}

func New(args ...any) error {
	var msg string
	var code codes.Code
	var wrapped error
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			if msg != "" {
				panic(fmt.Sprintf("errors.New provided unexpected second message argument %q", v))
			}
			msg = v
		case codes.Code:
			if code != codes.Internal {
				panic(fmt.Sprintf("errors.New provided unexpected second code argument %s", v))
			}
			code = v
		case error:
			if wrapped != nil {
				panic(fmt.Sprintf("errors.New provided unexpected second wrapped error argument %q", v))
			}
			wrapped = v
		default:
			panic(fmt.Sprintf("errors.New provided unexpected value %#v of type %T", v, v))
		}
	}

	if wrapped != nil && msg == "" && code == codes.Internal {
		panic("errors.New cannot be provided a wrapped error on its own")
	}

	return Error{
		msg:     msg,
		code:    code,
		wrapped: wrapped,
	}
}

func (e Error) Unwrap() error {
	return e.wrapped
}

func (e Error) Error() string {
	var parts []string
	if e.msg != "" {
		parts = append(parts, fmt.Sprintf("msg: %s", e.msg))
	}
	parts = append(parts, fmt.Sprintf("code: %s", e.code))
	if e.wrapped != nil {
		parts = append(parts, fmt.Sprintf("wrapped: %s", e.wrapped))
	}
	return strings.Join(parts, ", ")
}

func Message(err error) string {
	var urlshortErr Error
	if !errors.As(err, &urlshortErr) {
		return ""
	}
	if urlshortErr.msg != "" {
		return urlshortErr.msg
	} else {
		return Message(urlshortErr.wrapped)
	}
}

func Code(err error) codes.Code {
	var urlshortErr Error
	if !errors.As(err, &urlshortErr) {
		return codes.Internal
	}
	if urlshortErr.code != codes.Internal {
		return urlshortErr.code
	} else {
		return Code(urlshortErr.wrapped)
	}
}

var Is = errors.Is
var As = errors.As
var Unwrap = errors.Unwrap
