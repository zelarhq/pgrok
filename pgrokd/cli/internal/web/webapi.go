package web

import (
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/flamego/flamego"
	"unknwon.dev/x/anyx"
)

func getUserInfo(c *Context, r flamego.Render) {
	r.JSON(http.StatusOK, map[string]any{
		"displayName": c.Principal.DisplayName,
		"token":       c.Principal.Token,
		"url":         c.Config.Proxy.Scheme + "://" + c.Principal.Subdomain + "." + c.Config.Proxy.Domain,
		// Connection details so the dashboard can render copy-paste CLI commands.
		"subdomain":    c.Principal.Subdomain,
		"domain":       c.Config.Proxy.Domain,
		"scheme":       c.Config.Proxy.Scheme,
		"remoteAddr":   hostname(c.Config.ExternalURL) + ":" + strconv.Itoa(c.Config.SSHD.Port),
		"tcpHost":      hostname(anyx.Coalesce(c.Config.Proxy.TCP.Domain, c.Config.Proxy.Domain)),
		"tcpPortStart": c.Config.Proxy.TCP.PortStart,
		"tcpPortEnd":   c.Config.Proxy.TCP.PortEnd,
	})
}

// hostname returns the bare host of a URL or "host[:port]" string.
func hostname(s string) string {
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		return u.Hostname()
	}
	host, _, err := net.SplitHostPort(s)
	if err != nil {
		return s
	}
	return host
}

// getIdentityProvider reports the configured identity provider so the sign-in
// page can render the correct "Continue with ..." action.
func getIdentityProvider(c *Context, r flamego.Render) {
	if c.Config.IdentityProvider == nil {
		r.JSON(http.StatusInternalServerError, map[string]string{
			"error": "No identity provider is configured, please ask your admin to configure an identity provider.",
		})
		return
	}
	r.JSON(http.StatusOK, map[string]string{
		"displayName": c.Config.IdentityProvider.DisplayName,
		"authURL":     "/-/oidc/auth",
	})
}
