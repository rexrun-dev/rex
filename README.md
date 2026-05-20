<p align="center">
  <img src="assets/logo.png" alt="rex" width="320" />
</p>

<h3 align="center">run anything, know nothing.</h3>

<p align="center">
  <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go"></a>
  <a href="https://github.com/rexrun-dev/rex/actions"><img src="https://github.com/rexrun-dev/rex/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/rexrun-dev/rex"><img src="https://goreportcard.com/badge/github.com/rexrun-dev/rex" alt="Go Report Card"></a>
  <a href="https://app.codecov.io/gh/rexrun-dev/rex"><img src="https://codecov.io/gh/rexrun-dev/rex/branch/main/graph/badge.svg" alt="Coverage"></a>
  <a href="https://github.com/rexrun-dev/rex/releases"><img src="https://img.shields.io/github/v/release/rexrun-dev/rex" alt="Release"></a>
  <a href="https://github.com/rexrun-dev/rex/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue" alt="License"></a>
</p>

<p align="center">
  <a href="#install">Install</a> •
  <a href="#demo">Demo</a> •
  <a href="#commands">Commands</a> •
  <a href="#supported-stacks">Stacks</a> •
  <a href="#how-it-works">How it works</a>
</p>

---

**rex** detects your project's stack and runs the right command. No config file. No README reading. No memorizing.

```bash
git clone https://github.com/someone/unknown-project
cd unknown-project
rex test
```

That's it. Works for Go, Node, Python, Rust, Java, Docker, Make, and Just projects.

## Demo

https://github.com/rexrun-dev/rex/releases/download/v0.2.0/demo.mp4

<details>
<summary>Text version</summary>

```
$ cd some-go-project/
$ rex

  🦖 some-go-project

  stack  go

  commands
    rex test   → go test ./...
    rex run    → go run ./cmd/server
    rex build  → go build ./...
    rex deps   → go mod download
    rex fmt    → gofmt -w .
    rex lint   → go vet ./...
    rex clean  → go clean -cache

$ rex test
→ go test ./...
ok   github.com/example/app   1.2s
```

```
$ cd some-node-project/
$ rex

  🦖 some-node-project

  stack  node + pnpm

  commands
    rex test   → pnpm test
    rex run    → pnpm run dev
    rex build  → pnpm run build
    rex deps   → pnpm install
    rex clean  → rm -rf node_modules dist .next .nuxt build out
```

</details>

## Install

```bash
go install rexrun.dev/rex/cmd/rex@latest
```

Or download a binary from [Releases](https://github.com/rexrun-dev/rex/releases).

## Commands

| Command | What it does |
|---------|-------------|
| `rex` | Show project overview |
| `rex test` | Run tests |
| `rex run` | Start the app / dev server |
| `rex build` | Build the project |
| `rex deps` | Install dependencies |
| `rex clean` | Remove build artifacts |
| `rex fresh` | clean → deps → build (the "I pulled and it broke" fix) |
| `rex fmt` | Format code |
| `rex lint` | Lint code |
| `rex clone <url>` | Clone + detect + install deps (URL to ready in one command) |
| `rex doctor` | Diagnose environment issues |

### Flags

```bash
rex --list       # show all detected commands
rex --dry-run test  # show what would run without executing
rex test -- -v -run TestFoo  # pass extra args to underlying command
```

## Supported Stacks

| Stack | Package managers | Detection |
|-------|-----------------|-----------|
| **Go** | go modules | `go.mod` |
| **Node** | npm, pnpm, yarn, bun | `package.json` + lockfile |
| **Python** | pip, uv, poetry, pipenv | `pyproject.toml`, `requirements.txt`, `setup.py` |
| **Rust** | cargo | `Cargo.toml` |
| **Docker** | docker compose | `docker-compose.yml`, `compose.yml` |
| **Make** | — | `Makefile` |
| **Just** | — | `Justfile` |

Rex reads existing task runners first. If your project has a Makefile, Justfile, or Taskfile, those take priority.

## How It Works

1. **Detect** — Rex looks at your project root for known files (`go.mod`, `package.json`, `Cargo.toml`, etc.)
2. **Resolve** — It maps standard verbs (test, run, build...) to the correct command for your stack
3. **Execute** — It runs the command with full stdio passthrough and correct exit codes

No config needed. No lock-in. No magic.

### Priority chain

```
1. Justfile / Makefile (task runner overrides)
2. package.json scripts / Cargo.toml / go.mod (ecosystem)
3. Language heuristics (fallback)
```

If a Makefile defines a `test` target, `rex test` runs `make test` — even in a Go project. Rex respects what's already there.

## Philosophy

- **Zero config** — works out of the box, in any repo
- **Transparent** — always shows what it will run (the `→` line)
- **Fast** — single binary, <50ms startup, no network calls
- **Composable** — correct exit codes, works in CI pipelines
- **Non-invasive** — creates no files, modifies nothing

## Optional: `rex.toml`

For repos that want to pin commands explicitly:

```toml
[commands]
test = "go test -race -count=1 ./..."
run = "air"
build = "goreleaser build --snapshot"
```

Rex uses this first when present. Your team always gets the right command.

## CI Usage

```yaml
# GitHub Actions
- run: rex test

# That's it. Works for any stack.
```

## Contributing

```bash
git clone https://github.com/rexrun-dev/rex
cd rex
rex test  # yes, rex tests itself 🦖
```

## License

Apache 2.0
