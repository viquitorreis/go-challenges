package errors

import "errors"

var (
	ErrReadingFromConn = errors.New("err reading from conn")
	ErrWritingToConn   = errors.New("err writing to conn")
)
