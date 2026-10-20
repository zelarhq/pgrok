# pgrok — Zelar fork

Upstream [pgrok/pgrok](https://github.com/pgrok/pgrok) plus HTTP basic auth on tunnels.
Runs at `tunnel.zelar.com` (deployed from `zelarhq/openops` → `kustomize/pgrok`).

## Basic auth

```bash
pgrok http 3000 --basic-auth 'user:password'      # or -a, PGROK_BASIC_AUTH, or `basic_auth:` in pgrok.yml
pgrok init ... --basic-auth 'user:password'       # saves it to pgrok.yml for every run
```

- HTTP tunnels only; password ≥ 8 chars. Credential lives in pgrokd memory for the session only (never in the DB).
- pgrokd answers 401 + `WWW-Authenticate` without it, and strips `Authorization` before forwarding to your app.
- Fail-closed: the client refuses to go live unless the server confirms (`server-info` → `basic_auth: enabled`), so a stock pgrokd can't silently expose the tunnel.
- Needs **both** the fork client and the fork pgrokd. Without `--basic-auth` the fork is wire-compatible with upstream.

## Install as a service

The dashboard shows a one-liner per OS that downloads the client, writes the config and registers a service:

```bash
curl -fsSL https://tunnel.zelar.com/install.sh | sh -s -- --token <token> --forward-addr 3000 [--basic-auth user:pass]
```
```powershell
# Administrator PowerShell
& ([scriptblock]::Create((irm https://tunnel.zelar.com/install.ps1))) -Token <token> -ForwardAddr 3000 [-BasicAuth user:pass]
```

pgrokd serves both scripts (`pgrokd/cli/internal/web/install/`) with its own SSH address filled in. They wrap
`pgrok service install|uninstall|start|stop|restart|status`, which always runs `pgrok http` with the given config:

| OS | Service | Privileges | Logs |
|----|---------|------------|------|
| macOS | per-user LaunchAgent (starts at login) | none | `~/Library/Logs/pgrok.err.log` |
| Linux | systemd unit, `User=$SUDO_USER`, `Restart=always`, `RestartSec=5` | sudo | `journalctl -u pgrok -f` |
| Windows | Windows service, auto start, restart on failure | Administrator | `%USERPROFILE%\.pgrok\pgrok.log` |

The Linux unit is our own template (the library default waits 120s between restarts and gives up after 10).
`windows-service-test` in `.github/workflows/zelar.yml` runs the real `install.ps1` on Windows and gates the client release.

## Branches & releases

- `main` — mirror of upstream `main`, never committed to.
- `zelar` — upstream release tag + the fork commits. Default branch.
- Tag `vX.Y.Z-zelar.N` → `.github/workflows/zelar.yml` publishes `ghcr.io/zelarhq/pgrokd:X.Y.Z-zelar.N` and client binaries on the release.

## Rebasing on a new upstream release

```bash
git fetch upstream --tags
git checkout zelar && git rebase --onto vNEW vOLD zelar
go test ./internal/... && git push --force-with-lease origin zelar
git tag vNEW-zelar.1 && git push origin vNEW-zelar.1
```

Then bump the image tag in `openops/kustomize/pgrok/deployment.yml`.
