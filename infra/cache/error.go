package cache

type KeyNotExistsError struct {
	Key string
}

func (e *KeyNotExistsError) Error() string {
	return "cache key " + e.Key + " not exists"
}
