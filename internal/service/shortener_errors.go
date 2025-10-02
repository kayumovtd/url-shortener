package service

import (
	"fmt"
)

type ErrShortenerConflict struct {
	ResultURL string
	Err       error
}

func (e *ErrShortenerConflict) Error() string {
	return fmt.Sprintf("conflict: short URL already exists %q: %v", e.ResultURL, e.Err)
}

func NewErrShortenerConflict(resultURL string, err error) error {
	return &ErrShortenerConflict{ResultURL: resultURL, Err: err}
}
