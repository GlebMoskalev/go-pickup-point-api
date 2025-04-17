package repoerr

import "errors"

var (
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrNotFound       = errors.New("not found")
)
