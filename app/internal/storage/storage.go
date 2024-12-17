// Custom errors
package storage

import "errors"

var (
	ErrUidNotFound = errors.New("uid not found")
	ErrNoRecords   = errors.New("no records")
)
