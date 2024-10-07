package sys

import (
	"github.com/pkg/errors"
	"github.com/t34-dev/go-utils/pkg/sys/codes"
)

type commonError struct {
	msg  string
	code codes.Code
}

func NewError(msg string, code codes.Code) *commonError {
	return &commonError{msg, code}
}

func (r *commonError) Error() string {
	return r.msg
}

func (r *commonError) Code() codes.Code {
	return r.code
}

func IsError(err error) bool {
	var ce *commonError
	return errors.As(err, &ce)
}

func GetError(err error) *commonError {
	var ce *commonError
	if !errors.As(err, &ce) {
		return nil
	}

	return ce
}
