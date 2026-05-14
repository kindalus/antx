# Codebase Summary: antx

## Purpose
`antx` is an interactive, shell-like Go CLI for browsing Antbox nodes, managing files/folders/agents, and chatting with Antbox agents from a terminal.

## Architecture
This is a single-binary CLI. `main.go` enters Cobra command setup in `cmd/`, then `cli/` starts the interactive prompt and dispatches registered shell commands. API calls are isolated behind the `antbox.Antbox` interface and implemented by `antbox/client.go` against the Antbox HTTP API described by `openapi.yaml`.

## Key Modules
| Module / Path | Responsibility | Key Dependencies |
|---|---|---|
| `main.go` | Binary entry point; calls `cmd.Execute()`. | `cmd` |
| `cmd/root.go` | Top-level Cobra command, startup flags, auth mode selection. | `spf13/cobra`, `cli.StartOptions` |
| `cli/prompt.go` | Interactive prompt loop, command dispatch, agent shortcut parsing, startup initialization/cache restore. | `go-prompt`, `antbox` |
| `cli/command.go` + `cli/*` commands | Command interface and command implementations registered via `init()`. | shared `cli` globals, `antbox.Antbox` |
| `cli/lightray_auth.go` | Lightray OAuth Device Code Flow for `--lightray`; converts Lightray URL to `${url}/api`. | `net/http`, OIDC discovery/device/token endpoints |
| `cli/config.go` | Persists current node and last 20 commands in `~/.antx`. | filesystem |
| `antbox/antbox.go` | Public client interface and `NewClient` constructor. | `net/http` |
| `antbox/client.go` | HTTP implementation for nodes, files, agents, users, groups, docs, audit, auth. | `net/http`, `mimetype` |
| `antbox/types.go` | API DTOs/types for nodes, filters, agents, chat, users, groups, audit, errors. | stdlib |
| `openapi.yaml` | Source of truth for supported Antbox API shapes. | OpenAPI 3.1 |

## Data Flow
Startup: `main.go` → `cmd/root.go` parses `[server url]` and flags → `cli.StartWithOptions` builds an `antbox.Client` → optional root login or Lightray device flow → initialize root/current node, cached agents, `~/.antx` state → run `go-prompt`.

Command execution: user input → `cli.executor` → agent shortcuts (`/[uuid]`, `@[uuid]`) or registered `Command.Execute` → `antbox.Antbox` interface → HTTP request → terminal output and local state/cache updates.

## Build / Run / Test
| Action | Command |
|---|---|
| Build | `make build` or `go build -o antx` |
| Test | `make test` or `go test ./...` |
| Run with API key | `go run . <antbox-api-url> --api-key <key>` |
| Run with JWT | `go run . <antbox-api-url> --jwt <token>` |
| Run with root password | `go run . <antbox-api-url> --root <password>` |
| Run through Lightray | `go run . <lightray-url> --lightray` |

## Stack
- Language: Go `1.25.0` (`go.mod`).
- CLI framework: `github.com/spf13/cobra`.
- Interactive prompt/completion: `github.com/c-bata/go-prompt`.
- MIME detection: `github.com/gabriel-vasile/mimetype`.
- Markdown terminal rendering: `go.xrstf.de/go-term-markdown`.
- Tests: standard Go `testing`, `httptest`.

## Current Commands
Registered shell commands include: `agents`, `aliases`, `audit`, `cd`, `clone`, `cp`, `docs`, `download`, `exit`, `find`, `help`, `history`, `ls`, `mkdir`, `mksmart`, `mv`, `pwd`, `reload`, `rename`, `rm`, `sessions`, `stat`, `status`, `upload`, `whoami`.

Agent interactions are shortcuts, not registered commands:
- `/<agent_uuid> [message]` or `/[agent_uuid] [message]` for interactive chat.
- `@<agent_uuid> <question>` or `@[agent_uuid] <question>` for one-shot answer.

## Watch-outs
- Keep supported features aligned with `openapi.yaml`; unsupported legacy areas were intentionally removed: actions, templates, extensions, ai-tools, rag, aspects, and `duplicate` command.
- `clone` is the public command/method, but the server endpoint is still `/nodes/{uuid}/-/duplicate`.
- `find` accepts `~=` in CLI syntax but normalizes it to OpenAPI operator `match`.
- `audit` must call `GetNode(uuid)` first because `GetAuditLog` requires the node mimetype query parameter.
- Audit output hides payload/sensitive fields by default; only `audit -v` / `audit --verbose` prints payload.
- Lightray auth uses Device Code Flow with default `client_id=terminal-cli`, scope `openid profile email`, and sends the returned access token as `Authorization: Bearer ...` to `${lightray-url}/api`.
- Lightray waiting can be aborted with `Ctrl+C` or `Ctrl+D`.
- Global CLI state lives in package variables in `cli/prompt.go`; tests often replace `client` with mocks and reset globals manually.
- Persistent local CLI state is `~/.antx`; avoid changing this format without migration/backward compatibility.
- New CLI commands should implement `Command`, register in `init()`, update tests in `cli/prompt_test.go`, and avoid adding unsupported Antbox API areas.
- Run `go test ./...` before committing changes.
