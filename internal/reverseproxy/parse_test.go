package reverseproxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBasicAuth(t *testing.T) {
	got, err := ParseBasicAuth("alice:pa:ss:word")
	require.NoError(t, err)
	assert.Equal(t, BasicAuth{Username: "alice", Password: "pa:ss:word"}, got)

	for _, bad := range []string{"alice", ":password123", "alice:short", "test:test"} {
		_, err := ParseBasicAuth(bad)
		assert.Error(t, err, bad)
	}
}
