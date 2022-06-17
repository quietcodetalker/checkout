package stock

import "errors"

var (
	ErrInternal           = errors.New("internal")
	ErrNotFound           = errors.New("not found")
	ErrNotEnough          = errors.New("not enough")
	ErrInvalidMsg         = errors.New("invalid message")
	ErrFailedPrecondition = errors.New("failed precondition")
)
