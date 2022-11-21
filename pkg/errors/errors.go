package errors

import "errors"

var (
	ErrNeedRegistry        = errors.New("need registry address")
	ErrWrongRegistryFormat = errors.New("wrong registry address format")
	ErrNeedRepo            = errors.New("need repository")
	ErrNeedBlobId          = errors.New("need blob id")
	ErrNeedTagOrManifest   = errors.New("need tag or manifest")
	ErrTooManyArgs         = errors.New("too many args")
	ErrSplitAuth           = errors.New("can not split username and password")
	ErrNeedStdOut          = errors.New("need stdout")
	ErrUnknownOutput       = errors.New("unknown output format")
	ErrConflictAllRepo     = errors.New("confilct with all repositories")
	ErrUnknownManifest     = errors.New("unknown manifest")
)
