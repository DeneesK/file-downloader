package storage

import "errors"

var ErrNotFound = errors.New("a record with this key not found")
var ErrNotUniqueVallation = errors.New("a record with this key already exists")
