package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostname(t *testing.T) {
	for in, want := range map[string]string{
		"https://tunnel.example.com":       "tunnel.example.com",
		"http://localhost:3320":            "localhost",
		"tunnel.example.com":               "tunnel.example.com",
		"localhost:3000":                   "localhost",
		"https://tunnel.example.com:8443/": "tunnel.example.com",
	} {
		assert.Equal(t, want, hostname(in), in)
	}
}
