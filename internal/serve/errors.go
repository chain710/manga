package serve

import "errors"

var (
	errInvalidRequest = errors.New("invalid request")
	errNotFound       = errors.New("not found resource")
	errScanDisabled   = errors.New("scan disabled")
)
