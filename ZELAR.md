# pgrok — Zelar fork

Upstream [pgrok/pgrok](https://github.com/pgrok/pgrok) plus HTTP basic auth on tunnels.
Runs at `tunnel.zelar.com` (deployed from `zelarhq/openops` → `kustomize/pgrok`).

## Basic auth

```bash
pgrok http 3000 --basic-auth 'user:password'      # or -a, PGROK_BASIC_AUTH, or `basic_auth:` in pgrok.yml
```

- HTTP tunnels only; password ≥ 8 chars. Credential lives in pgrokd memory for the session only (never in the DB).
- pgrokd answers 401 + `WWW-Authenticate` without it, and strips `Authorization` before forwarding to your app.
- Fail-closed: the client refuses to go live unless the server confirms (`server-info` → `basic_auth: enabled`), so a stock pgrokd can't silently expose the tunnel.
- Needs **both** the fork client and the fork pgrokd. Without `--basic-auth` the fork is wire-compatible with upstream.

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
