package sshd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pgrok/pgrok/internal/reverseproxy"
)

func TestParseBasicAuth(t *testing.T) {
	got, err := parseBasicAuth("alice:pa:ss:word")
	require.NoError(t, err)
	assert.Equal(t, reverseproxy.BasicAuth{Username: "alice", Password: "pa:ss:word"}, got)

	for _, bad := range []string{"alice", ":password123", "alice:short"} {
		_, err := parseBasicAuth(bad)
		assert.Error(t, err, bad)
	}
}
