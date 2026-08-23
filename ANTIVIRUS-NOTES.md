# Antivirus / heuristic detection notes

The Windows release binary occasionally gets flagged by antivirus engines on
VirusTotal — mostly low-reputation heuristic/ML engines, sometimes Microsoft
Defender's `!ml` heuristic model. This is expected for this kind of binary
and, as far as we can tell, a false positive. This document explains why,
and points at the exact code responsible, so it's easy to re-evaluate if
that ever changes.

## Why this happens at all

Every release binary starts from zero reputation: newly built, unsigned, and
cross-compiled for Windows from a Linux CI runner (see
`.github/workflows/release.yml`). Combined with `-ldflags="-s -w" -trimpath`
(stripped symbols, no build paths — see the `Makefile`), the binary has
higher entropy than a typical unstripped executable, which nudges
entropy-based heuristics toward "packed." None of that is specific to this
project — it's the standard false-positive profile for small, unsigned,
freshly-built Go binaries, and it generally clears up once the binary/hash
accumulates reputation or gets code-signed.

What's more interesting, and worth documenting, is that a few of this
project's *legitimate* features have a shape that overlaps with what
heuristic and behavioral engines are specifically built to catch.

## Specific functionality that looks suspicious to heuristics

### 1. Fail-fast on missing environment variables

[`cmd/paperless-ngx-mcp/main.go:15-25`](cmd/paperless-ngx-mcp/main.go#L15-L25)

```go
paperlessURL := os.Getenv("PAPERLESS_URL")
if paperlessURL == "" {
    fmt.Fprintln(os.Stderr, "PAPERLESS_URL is required")
    os.Exit(1)
}

paperlessToken := os.Getenv("PAPERLESS_TOKEN")
if paperlessToken == "" {
    fmt.Fprintln(os.Stderr, "PAPERLESS_TOKEN is required")
    os.Exit(1)
}
```

The server requires `PAPERLESS_URL` and `PAPERLESS_TOKEN` to be set, and
exits immediately if they aren't. That's ordinary configuration validation,
but "check for specific environment variables and bail out instantly if
they're absent" is also a well-known anti-sandbox evasion technique —
malware uses it to detect analysis sandboxes, which rarely set
application-specific environment variables, and exit before doing anything
observable. A behavioral scanner running the raw binary with no environment
configured will see exactly that pattern: process starts, does nothing
visible, exits in milliseconds. This is likely the single biggest
contributor to `/behavior`-tab flags on VirusTotal.

There's no fix here beyond documenting it — requiring real config before
talking to a real Paperless instance is correct behavior, not evasion.

### 2. Arbitrary local file → network upload

[`internal/tools/upload_document.go`](internal/tools/upload_document.go)

- Opens a caller-specified file: [`os.Open(params.FilePath)` at line 115](internal/tools/upload_document.go#L115)
- Streams its contents into a multipart body and POSTs it to the configured
  Paperless server: [`t.client.PostMultipart(...)` at lines 203-208](internal/tools/upload_document.go#L203-L208)

Read-arbitrary-local-file-then-send-over-HTTP is the generic shape of an
info-stealer's exfiltration step. Here it's the intended feature — the tool
uploads a document you specify to the Paperless server you configured, using
the token you provided — but static analysis has no way to know that; it
just sees a file read followed by a network POST.

### 3. Arbitrary network content → local file write

[`internal/tools/download_document.go`](internal/tools/download_document.go)

- Fetches document content from the configured Paperless server:
  [`t.client.Get(ctx, path)` at line 103](internal/tools/download_document.go#L103)
- Streams the response body to a caller-specified path:
  [`os.Create(params.SavePath)` at line 121](internal/tools/download_document.go#L121)

The mirror image of (2), and the generic shape of a dropper writing
attacker-controlled content to an arbitrary path. Path safety here is
handled by `validateFilePath` (see below) — the path must be absolute and
must not contain `..` components — but again, that nuance isn't visible to
a heuristic scanner looking only at API call shape.

Path validation for both tools lives in
[`internal/tools/helpers.go:303-326`](internal/tools/helpers.go#L303-L326)
(`validateFilePath`): it rejects any `..` path component before cleaning
(since `filepath.Clean` would otherwise silently resolve traversal
sequences) and requires the cleaned path to be absolute. This prevents
traversal but doesn't sandbox the destination to any particular directory —
by design, since the caller is expected to choose where documents land.

### 4. The MCP command-loop shape, generally

[`internal/server/server.go:190-200`](internal/server/server.go#L190-L200)
(`Server.Run`)

The server reads newline-delimited JSON-RPC requests from stdin, dispatches
each one to a named handler (a "tool") based on the request, and writes a
JSON response to stdout. That's the entire protocol — and it's also the
generic shape of a C2 implant's command loop: receive a command, act on the
filesystem/network based on it, report the result back over a pipe. This
isn't specific to anything this project does; it's inherent to what an MCP
server *is*. Expect any MCP server implementation, not just this one, to run
somewhat hotter on heuristic engines than an equivalent non-agentic CLI
tool.

## What we did *not* find

No `os/exec`, no `syscall`, no `unsafe`, no `encoding/base64` or other
string obfuscation, no listening sockets, no registry access, no
persistence mechanism (scheduled tasks, startup entries, etc.), and no
reflection-based code execution. Every code path traces directly to a
documented MCP tool call — there is no hidden or conditionally-triggered
behavior anywhere in the source.

## If a detection needs to be taken seriously

The false-positive read above holds as long as:

- The flagging engines are low-reputation/heuristic (`Gen:Variant.*`,
  `ML/Attribute-HG`, `Trojan.GenericKD`, Defender's `!ml` suffix, etc.),
  not a specific named family with signature-based detection across
  multiple major engines.
- The binary was built by this repo's own `release.yml` workflow from a tag
  on `master`, not downloaded from an untrusted mirror.

If either of those stop being true, treat it as a real finding, not
noise — re-diff the release commit against `upstream/master` and check the
Actions run logs for that specific build.

## Reducing false positives going forward

- **Code-signing** is the actual fix — publisher reputation is the dominant
  signal most heuristic engines weight, and a signed binary from a
  consistent publisher will stop tripping cold-start heuristics almost
  entirely. Costs money (a code-signing certificate); probably overkill for
  a personal-use binary, worth it if this gets wider distribution.
- **Submitting flagged builds to vendor false-positive review** (e.g.
  Microsoft's at <https://www.microsoft.com/en-us/wdsi/filesubmission>)
  builds reputation for this hash/publisher over time and is free.
