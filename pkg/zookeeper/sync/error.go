package sync

import "errors"

var (
	ErrAcquireTimeout = errors.New("acquire lock timeout")
)