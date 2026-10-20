package reverseproxy

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
	"github.com/puzpuzpuz/xsync/v2"
)

// BasicAuth is the HTTP basic authentication credential a tunnel requires its
// visitors to present. The zero value means no authentication.
type BasicAuth struct {
	Username string
	Password string
}

// Cluster contains a list of proxies identified by their hosts.
type Cluster struct {
	proxies     map[string]http.Handler
	proxiesLock xsync.RBMutex
}

// NewCluster returns a new Cluster.
func NewCluster() *Cluster {
	return &Cluster{proxies: make(map[string]http.Handler)}
}

// Get returns the proxy by the given host. It returns a boolean to indicate
// whether such proxy exists.
func (c *Cluster) Get(host string) (http.Handler, bool) {
	t := c.proxiesLock.RLock()
	defer c.proxiesLock.RUnlock(t)

	proxy, ok := c.proxies[host]
	return proxy, ok
}

// Set atomically sets a proxy for the given host. If the host is already taken,
// it tries the alternative host. It returns the actually used host or an error
// if both are taken.
func (c *Cluster) Set(host, alternative, forward string, auth BasicAuth) (string, error) {
	c.proxiesLock.Lock()
	defer c.proxiesLock.Unlock()

	if _, exists := c.proxies[host]; !exists {
		c.proxies[host] = newProxy(forward, auth)
		return host, nil
	}
	if _, exists := c.proxies[alternative]; !exists {
		c.proxies[alternative] = newProxy(forward, auth)
		return alternative, nil
	}
	return "", errors.New("host and alternative are both taken")
}

func newProxy(forward string, auth BasicAuth) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = forward
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(errors.Cause(err).Error()))
		},
	}
	if auth.Username == "" {
		return proxy
	}

	// Compare digests so the comparison is constant-time regardless of input length.
	wantUser := sha256.Sum256([]byte(auth.Username))
	wantPass := sha256.Sum256([]byte(auth.Password))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		gotUser := sha256.Sum256([]byte(user))
		gotPass := sha256.Sum256([]byte(pass))
		if !ok ||
			subtle.ConstantTimeCompare(gotUser[:], wantUser[:])&
				subtle.ConstantTimeCompare(gotPass[:], wantPass[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="pgrok", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// The credential is for the tunnel, not for the application behind it.
		r.Header.Del("Authorization")
		proxy.ServeHTTP(w, r)
	})
}

// Remove removes the proxy with given host from the cluster.
func (c *Cluster) Remove(host string) {
	c.proxiesLock.Lock()
	defer c.proxiesLock.Unlock()
	delete(c.proxies, host)
}
