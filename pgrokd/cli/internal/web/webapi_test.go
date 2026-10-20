package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pgrok/pgrok/internal/conf"
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

func TestServeInstallScript(t *testing.T) {
	config := &conf.Config{ExternalURL: "https://tunnel.example.com"}
	config.SSHD.Port = 2222
	for _, script := range []string{installSh, installPs1} {
		w := httptest.NewRecorder()
		serveInstallScript(config, script, "text/plain")(w, httptest.NewRequest(http.MethodGet, "/", nil))
		body := w.Body.String()
		assert.Contains(t, body, `"tunnel.example.com:2222"`)
		assert.Contains(t, body, ClientDownloadURL)
		assert.NotContains(t, body, "__")
	}
}
