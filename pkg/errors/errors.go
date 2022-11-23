package errors

import "errors"

var (
	ErrWrongRegistryAddress = errors.New("wrong registry address format")
	ErrNeedTag              = errors.New("need tag")
	ErrNeedImageReference   = errors.New("need image reference")
	ErrTooManyArgs          = errors.New("too many args")
	ErrSplitAuth            = errors.New("can not split username and password")
	ErrNeedStdOut           = errors.New("need stdout")
	ErrUnknownOutput        = errors.New("unknown output format")
	ErrUnknownSort          = errors.New("unknown sort method")
	ErrUnknownManifest      = errors.New("unknown manifest")
)
