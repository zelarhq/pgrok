package web

import (
	_ "embed"
	"net/http"
	"strconv"
	"strings"

	"github.com/pgrok/pgrok/internal/conf"
)

// ClientDownloadURL is where the install scripts fetch client binaries from; the
// fork's release workflow publishes them on every v*-zelar.* tag.
const ClientDownloadURL = "https://github.com/zelarhq/pgrok/releases/latest/download"

var (
	//go:embed install/install.sh
	installSh string
	//go:embed install/install.ps1
	installPs1 string
)

// serveInstallScript serves a client install script with this server's
// addresses filled in, so users only pass their token.
func serveInstallScript(config *conf.Config, script, contentType string) http.HandlerFunc {
	body := strings.NewReplacer(
		"__SERVER_URL__", config.ExternalURL,
		"__REMOTE_ADDR__", hostname(config.ExternalURL)+":"+strconv.Itoa(config.SSHD.Port),
		"__DOWNLOAD_URL__", ClientDownloadURL,
	).Replace(script)
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write([]byte(body))
	}
}
