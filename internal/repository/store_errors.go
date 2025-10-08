package repository

import "fmt"

type ErrStoreConflict struct {
	ShortURL    string
	OriginalURL string
	Err         error
}

func (e *ErrStoreConflict) Error() string {
	return fmt.Sprintf("conflict: short URL %q already exists for original URL %q: %v", e.ShortURL, e.OriginalURL, e.Err)
}

func NewErrStoreConflict(shortURL, originalURL string, err error) *ErrStoreConflict {
	return &ErrStoreConflict{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		Err:         err,
	}
}
