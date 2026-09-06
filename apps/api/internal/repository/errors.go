package repository

import (
	"errors"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/lib/pq"
)

const (
	StatusPublished int16 = 0
	StatusHidden    int16 = 1
	StatusDeleted   int16 = 2
)

var (
	ErrInvalidTag     = errors.New("invalid or inactive tag")
	ErrCommentsLocked = errors.New("comments are locked")
)

func IsNotFound(err error) bool {
	return errors.Is(err, qrm.ErrNoRows)
}

func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
