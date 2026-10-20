import { Button } from "@/components/retroui/Button";
import { Card } from "@/components/retroui/Card";
import { Text } from "@/components/retroui/Text";
import { cn } from "@/lib/utils";
import { ReactNode, useState } from "react";
import CodeBlock from "../components/CodeBlock";
import CopyButton from "../components/CopyButton";
import ThemeToggle from "../components/ThemeToggle";
import useUser from "../hooks/useUser";

// Client binaries are published by the fork's release workflow; "latest" always
// tracks the server version deployed from the same fork.
const RELEASES_URL = "https://github.com/zelarhq/pgrok/releases";
const DOWNLOAD_URL = `${RELEASES_URL}/latest/download`;

type Platform = "macos" | "linux" | "windows";

const PLATFORMS: { id: Platform; label: string; arches: { id: string; label: string }[] }[] = [
  {
    id: "macos",
    label: "macOS",
    arches: [
      { id: "arm64", label: "Apple silicon" },
      { id: "amd64", label: "Intel" },
    ],
  },
  {
    id: "linux",
    label: "Linux",
    arches: [
      { id: "amd64", label: "x86_64" },
      { id: "arm64", label: "arm64" },
    ],
  },
  { id: "windows", label: "Windows", arches: [{ id: "amd64", label: "x86_64" }] },
];

function detectPlatform(): Platform {
  const ua = navigator.userAgent;
  if (/Mac/i.test(ua)) return "macos";
  if (/Win/i.test(ua)) return "windows";
  return "linux";
}

function installCommand(platform: Platform, arch: string): string {
  if (platform === "windows") {
    return [
      `Invoke-WebRequest -Uri ${DOWNLOAD_URL}/pgrok_windows_${arch}.exe -OutFile pgrok.exe`,
      String.raw`.\pgrok.exe --version`,
    ].join("\n");
  }
  const os = platform === "macos" ? "darwin" : "linux";
  return [
    `curl -fsSL -o pgrok \\\n  ${DOWNLOAD_URL}/pgrok_${os}_${arch}`,
    `chmod +x pgrok && sudo mv pgrok /usr/local/bin/pgrok`,
    `pgrok --version`,
  ].join("\n");
}

export default function DashboardPage() {
  const user = useUser();
  const [revealToken, setRevealToken] = useState(false);
  const [port, setPort] = useState("3000");
  const p = port || "3000";
  const bin = detectPlatform() === "windows" ? String.raw`.\pgrok.exe` : "pgrok";
  const urlFor = (prefix: string) => `${user.scheme}://${prefix}${user.subdomain}.${user.domain}`;

  return (
    <div className="min-h-screen">
      <nav className="border-b-2 border-border bg-card">
        <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-2">
            <img className="h-8 w-auto" src="/pgrok.png" alt="pgrok" />
            <Text as="h4">pgrok</Text>
          </div>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="link"
              className="hidden sm:flex"
              render={<a href={RELEASES_URL} target="_blank" rel="noreferrer" />}
            >
              Releases
            </Button>
            <ThemeToggle />
            <Button size="sm" variant="outline" render={<a href="/-/sign-out" />}>
              Sign out
            </Button>
          </div>
        </div>
      </nav>

      <main className="mx-auto max-w-5xl space-y-10 px-4 py-10 sm:px-6 lg:px-8">
        <header>
          <Text as="h1">Hi, {user.displayName}</Text>
          <Text className="mt-2 text-muted-foreground">Expose a local server to the internet in three steps.</Text>
        </header>

        {/* Account */}
        <Card className="block w-full">
          <Card.Content className="grid gap-6 md:grid-cols-2">
            <div className="min-w-0">
              <Label>Your public URL</Label>
              <div className="mt-2 flex items-center gap-2">
                <a
                  className="min-w-0 truncate font-head text-lg underline decoration-primary decoration-4 underline-offset-4"
                  href={user.url}
                  target="_blank"
                  rel="noreferrer"
                >
                  {user.url.replace(/^https?:\/\//, "")}
                </a>
                <CopyButton value={user.url} label="Copy URL" />
              </div>
              <Text className="mt-2 text-sm text-muted-foreground">Live while your pgrok client is connected.</Text>
            </div>
            <div className="min-w-0">
              <Label>Access token</Label>
              <div className="mt-2 flex items-center gap-2">
                <code className="min-w-0 flex-1 truncate rounded border-2 border-border bg-background px-2 py-1 font-mono text-sm">
                  {revealToken ? user.token : "•".repeat(24)}
                </code>
                <Button size="sm" variant="outline" onClick={() => setRevealToken(!revealToken)}>
                  {revealToken ? "Hide" : "Show"}
                </Button>
                <CopyButton value={user.token} label="Copy token" />
              </div>
              <Text className="mt-2 text-sm text-muted-foreground">
                Treat it like a password. Copied commands below include it.
              </Text>
            </div>
          </Card.Content>
        </Card>

        {/* Quick start */}
        <section>
          <Text as="h2">Get started</Text>
          <ol className="mt-4 space-y-4">
            <Step n={1} title="Install the CLI">
              <InstallTabs />
            </Step>
            <Step n={2} title="Save your connection settings">
              <Text className="mb-3 text-sm text-muted-foreground">
                Writes <code className="font-mono">~/.pgrok/pgrok.yml</code> so later commands need no flags.
              </Text>
              <CodeBlock
                code={`${bin} init --remote-addr ${user.remoteAddr} --forward-addr http://localhost:${p} --token ${user.token}`}
                secret={user.token}
                reveal={revealToken}
              />
            </Step>
            <Step n={3} title="Start a tunnel">
              <CodeBlock code={`${bin} http ${p}`} />
              <Text className="mt-3 text-sm">
                Then open <Link href={user.url}>{user.url}</Link>.
              </Text>
            </Step>
          </ol>
        </section>

        {/* Recipes */}
        <section>
          <div className="flex flex-wrap items-end justify-between gap-4">
            <Text as="h2">Usage</Text>
            <label className="flex items-center gap-2 text-sm font-semibold">
              Local port
              <input
                className="w-24 rounded border-2 border-border bg-card px-2 py-1 font-mono shadow-sm focus-visible:outline-2 focus-visible:outline-primary"
                inputMode="numeric"
                placeholder="3000"
                value={port}
                onChange={(event) => setPort(digitsOnly(event.target.value).slice(0, 5))}
              />
            </label>
          </div>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <Recipe title="Expose an HTTP server" result={user.url}>
              <CodeBlock code={`${bin} http ${p}`} />
            </Recipe>
            <Recipe title="Pick a fixed subdomain" result={urlFor("demo-")}>
              <CodeBlock code={`${bin} http ${p} --subdomain demo`} />
            </Recipe>
            <Recipe
              title="Protect it with a password"
              result="Password ≥ 8 chars. Pass --basic-auth to pgrok init to always require it."
            >
              <CodeBlock code={`${bin} http ${p} --basic-auth 'user:a-long-password'`} />
            </Recipe>
            <Recipe title="Expose a TCP port" result={`tcp://${user.tcpHost}:<${user.tcpPortStart}–${user.tcpPortEnd}>`}>
              <CodeBlock code={`${bin} tcp 5432`} />
            </Recipe>
            <Recipe title="Run without a config file" result="Handy for CI and one-off machines">
              <CodeBlock
                code={`${bin} http ${p} --remote-addr ${user.remoteAddr} --token ${user.token}`}
                secret={user.token}
                reveal={revealToken}
              />
            </Recipe>
            <Recipe title="Route paths to other ports" result="Add to ~/.pgrok/pgrok.yml">
              <CodeBlock code={"dynamic_forwards: |\n  /api http://localhost:8080\n  /ws  http://localhost:9000"} />
            </Recipe>
          </div>
        </section>

        {/* Troubleshooting */}
        <Card className="block w-full">
          <Card.Header className="border-b-2 border-border">
            <Card.Title className="mb-0">Troubleshooting</Card.Title>
          </Card.Header>
          <Card.Content>
            <dl className="divide-y-2 divide-border text-sm">
              <Faq q={`"query server info rejected" or "server does not support basic auth"`}>
                Your client doesn&apos;t match this server. Reinstall it with step 1.
              </Faq>
              <Faq q={`"no supported methods remain"`}>
                The token is wrong or outdated. Copy it again from this page and re-run step 2.
              </Faq>
              <Faq q="502 when opening the URL">
                Nothing is listening on <code className="font-mono">localhost:{p}</code>, or the client isn&apos;t running.
              </Faq>
              <Faq q="macOS says the binary can't be verified">
                Run <code className="font-mono">xattr -d com.apple.quarantine /usr/local/bin/pgrok</code> (only needed
                when downloaded with a browser).
              </Faq>
            </dl>
          </Card.Content>
        </Card>
      </main>
    </div>
  );
}

function InstallTabs() {
  const [platform, setPlatform] = useState<Platform>(detectPlatform);
  const current = PLATFORMS.find((x) => x.id === platform)!;
  const [arch, setArch] = useState(current.arches[0].id);

  return (
    <div>
      <div className="flex flex-wrap items-center gap-2" role="tablist" aria-label="Operating system">
        {PLATFORMS.map((x) => (
          <Segment
            key={x.id}
            active={x.id === platform}
            onClick={() => {
              setPlatform(x.id);
              setArch(x.arches[0].id);
            }}
          >
            {x.label}
          </Segment>
        ))}
        {current.arches.length > 1 && (
          <>
            <span className="mx-1 h-6 w-0.5 bg-border" aria-hidden="true" />
            {current.arches.map((a) => (
              <Segment key={a.id} small active={a.id === arch} onClick={() => setArch(a.id)}>
                {a.label}
              </Segment>
            ))}
          </>
        )}
      </div>
      <div className="mt-3">
        <CodeBlock code={installCommand(platform, arch)} />
      </div>
      <Text className="mt-2 text-sm text-muted-foreground">
        Checksums and all builds on the <Link href={RELEASES_URL}>releases page</Link>.
      </Text>
    </div>
  );
}

function Segment({
  active,
  small,
  onClick,
  children,
}: {
  active: boolean;
  small?: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "cursor-pointer rounded border-2 border-border font-head transition-all focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary",
        small ? "px-2 py-0.5 text-xs" : "px-3 py-1 text-sm",
        active ? "bg-primary text-primary-foreground shadow-sm" : "bg-card hover:bg-accent hover:text-accent-foreground",
      )}
    >
      {children}
    </button>
  );
}

function Step({ n, title, children }: { n: number; title: string; children: ReactNode }) {
  return (
    <li>
      <Card className="block w-full">
        <Card.Content className="flex gap-4">
          <span className="flex size-9 shrink-0 items-center justify-center rounded border-2 border-border bg-primary font-head text-primary-foreground">
            {n}
          </span>
          <div className="min-w-0 flex-1">
            <Text as="h5" className="mb-3">
              {title}
            </Text>
            {children}
          </div>
        </Card.Content>
      </Card>
    </li>
  );
}

function Recipe({ title, result, children }: { title: string; result: string; children: ReactNode }) {
  return (
    <Card className="flex w-full min-w-0 flex-col">
      <Card.Content className="flex flex-1 flex-col gap-3">
        <Text as="h6">{title}</Text>
        {children}
        <Text className="mt-auto truncate text-sm text-muted-foreground" title={result}>
          → {result}
        </Text>
      </Card.Content>
    </Card>
  );
}

function Faq({ q, children }: { q: string; children: ReactNode }) {
  return (
    <div className="grid gap-1 py-3 sm:grid-cols-5 sm:gap-4">
      <dt className="font-semibold sm:col-span-2">{q}</dt>
      <dd className="text-muted-foreground sm:col-span-3">{children}</dd>
    </div>
  );
}

function digitsOnly(s: string): string {
  return [...s].filter((c) => c >= "0" && c <= "9").join("");
}

function Label({ children }: { children: ReactNode }) {
  return <div className="text-xs font-semibold tracking-wider text-muted-foreground uppercase">{children}</div>;
}

function Link({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a className="underline decoration-primary underline-offset-2" href={href} target="_blank" rel="noreferrer">
      {children}
    </a>
  );
}
