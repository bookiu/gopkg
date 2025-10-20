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
			want: "cache key not exists. key=test-key",
		},
		{
			name: "empty key",
			key:  "",
			want: "cache key not exists. key=",
		},
		{
			name: "key with special characters",
			key:  "test:key:123",
			want: "cache key not exists. key=test:key:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newKeyNotExistsError(tt.key)
			assert.Equal(t, tt.want, e.Error())
		})
	}
}
