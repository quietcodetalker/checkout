package kafka

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInternal        = errors.New("internal")
)
