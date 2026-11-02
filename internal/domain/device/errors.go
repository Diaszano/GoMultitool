package device

import "errors"

// ErrInvalidDevice indicates incomplete, inconsistent, or unsupported metadata.
var ErrInvalidDevice = errors.New("invalid device metadata")
