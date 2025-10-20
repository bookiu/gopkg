package cache

import (
	"errors"
	"fmt"
)

var KeyNotExistsError = errors.New("cache key not exists")

func newKeyNotExistsError(k string) error {
	return fmt.Errorf("%w. key=%s", KeyNotExistsError, k)
}
