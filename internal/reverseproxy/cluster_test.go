package reverseproxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCluster_BasicAuth(t *testing.T) {
	var gotAuthorization string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()
	forward := strings.TrimPrefix(backend.URL, "http://")

	c := NewCluster()
	_, err := c.Set("open.example.com", "", forward, BasicAuth{})
	require.NoError(t, err)
	_, err = c.Set("locked.example.com", "", forward, BasicAuth{Username: "alice", Password: "s3cret-pass"})
	require.NoError(t, err)

	serve := func(host, user, pass string) *httptest.ResponseRecorder {
		proxy, ok := c.Get(host)
		require.True(t, ok)
		r := httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil)
		if user != "" || pass != "" {
			r.SetBasicAuth(user, pass)
		}
		w := httptest.NewRecorder()
		proxy.ServeHTTP(w, r)
		return w
	}

	t.Run("no auth configured", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, serve("open.example.com", "", "").Code)
	})
	t.Run("missing credential", func(t *testing.T) {
		w := serve("locked.example.com", "", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Basic")
	})
	t.Run("wrong password", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, serve("locked.example.com", "alice", "nope").Code)
	})
	t.Run("wrong username", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, serve("locked.example.com", "bob", "s3cret-pass").Code)
	})
	t.Run("correct credential is stripped before forwarding", func(t *testing.T) {
		w := serve("locked.example.com", "alice", "s3cret-pass")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, gotAuthorization)
	})
}
