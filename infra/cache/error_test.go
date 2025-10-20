package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyNotExistsError_Error(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "normal key",
			key:  "test-key",
			want: "cache key test-key not exists",
		},
		{
			name: "empty key",
			key:  "",
			want: "cache key  not exists",
		},
		{
			name: "key with special characters",
			key:  "test:key:123",
			want: "cache key test:key:123 not exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &KeyNotExistsError{
				Key: tt.key,
			}
			assert.Equal(t, tt.want, e.Error())
		})
	}
}
