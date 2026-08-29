# Cool-Code: a fast, native CLI coding agent

An intelligent command-line coding agent - describe what you want, and it reads, plans, and edits your codebase using live tools. Similar in spirit to Claude Code and Gemini CLI, now rewritten in **Go** with a **Bubble Tea** TUI and **native provider tool-calling**.

> **Rewritten in Go.** Cool-Code was previously a TypeScript/Node CLI. It is now a single statically-linked Go binary - no `node_modules`, faster start-up, real goroutine concurrency, and a polished Charm-stack terminal UI. The agent uses each provider's native function-calling API (Anthropic, OpenAI, Gemini) instead of parsing tool calls out of free text.

## Overview

Cool-Code combines large language models with a comprehensive set of development tools for an interactive development experience. It works with multiple model providers (Google, OpenAI, Anthropic) and finds context live through search and read tools - no vector database required.

## Features

- **Multiple model providers** - Google, OpenAI, or Anthropic. Connect one with `/connect` (keys stored in `~/.coolcode/credentials.json`) or export the matching env var; the provider is inferred from the model id.
- **Native tool-calling with streaming** - structured function-calling against each provider's API. Transport responses stream internally, while only completed final answers enter the transcript.
- **Explore subagents** - the agent can fan out several read-only `spawn_agent` explorers concurrently to investigate independent areas of a codebase, with live per-agent status in the TUI.
- **Concurrent tools + cancellation** - independent read-only tool calls run in parallel; Esc or Ctrl+C cancels a running turn without quitting.
- **Polished TUI** - a Bubble Tea terminal UI with Glamour-rendered markdown, multiline input (Alt+Enter), mouse-wheel scrolling, input-history recall, colored diff previews for edits, a `/` slash-command palette with Tab autocomplete, a live task panel, and Shift+Tab to switch modes.
- **Three agent modes** - Plan (read-only investigation → detailed plan), Agent (autonomous execution), Ask (read-only Q&A). After Plan mode produces a plan, choose **Start implementation** to jump straight into Agent mode.
- **Project memory (`COOLCODE.md`)** - persistent project instructions loaded into every prompt.
- **Skills** - discoverable, model-invoked instruction modules under `.coolcode/skills/` (compatible with Claude Code skills).
- **Web access** - `web_fetch` and `web_search` tools.
- **Session persistence** - conversations saved to `~/.coolcode/sessions/` (including on quit and cancel); resume with `--continue` / `--resume` and the prior conversation reappears in the transcript.
- **Reliability** - automatic retry with backoff on transient API errors, and real token usage from provider responses in the status bar.
- **Task tracking, input queuing, safety guardrails** - real-time checklists, mid-turn message queuing, and path/read/danger protections.

## Install

### Go install (recommended)

```bash
go install github.com/rushikeshg25/cool-code@latest
```

This places a `cool-code` binary in your `$GOBIN` (usually `~/go/bin` - make sure it's on your `PATH`).

### Build from source

```bash
git clone https://github.com/rushikeshg25/cool-code.git
cd cool-code
make build      # produces ./cool-code
# or: go build -o cool-code .
```

### Connect a provider

The easiest way: start `cool-code` and run **`/connect`** - pick a provider, paste your API key, and it is stored in `~/.coolcode/credentials.json` (mode 0600). The chosen provider and default model are saved in `~/.coolcode/settings.json`. Repository configuration cannot override provider identity, endpoints, credential selection, guardrails, or confirmation policy.

Env vars keep working as a fallback:

```bash
export GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here   # Gemini (default)
# export OPENAI_API_KEY=your_api_key_here               # OpenAI
# export ANTHROPIC_API_KEY=your_api_key_here            # Anthropic
```

OpenAI-compatible gateways and CLI proxies must be configured in trusted global settings. The CLI routes security-sensitive keys to `~/.coolcode/settings.json` automatically:

```bash
cool-code config set llm.provider openai
cool-code config set llm.model gpt-5.6-sol
cool-code config set llm.baseUrl http://localhost:8317/v1
cool-code config set llm.apiKeyEnv CLIPROXY_API_KEY
export CLIPROXY_API_KEY=your_proxy_key
```

`COOLCODE_API_BASE_URL` and `COOLCODE_API_KEY` provide equivalent process-environment overrides. Custom endpoints require an explicit proxy key; first-party provider credentials are never forwarded to them. Remote endpoints require HTTPS, while HTTP is accepted only for loopback proxies. A trusted LAN proxy can be explicitly enabled with `cool-code config set llm.allowInsecureHttp true`; this weakens transport security and should not be used across untrusted networks.

A regular, non-symlink `.env` file may supply variables ending in `_API_KEY`; point `llm.apiKeyEnv` at one to use it. `COOLCODE_API_KEY` is **not** read from `.env` - it takes effect with no global setting, so a cloned repository could otherwise substitute the credential your requests are billed and logged against. Export it in your shell instead. Endpoint and process-control variables in project `.env` files are ignored.

## Usage

Run inside your project directory:

```bash
cool-code
```

### Command-line options

```bash
cool-code --yes               # start without the init prompt
cool-code --no-init           # exit without initializing
cool-code --allow-dangerous   # skip danger confirmations
cool-code --copy              # copy final responses to the clipboard
cool-code --continue          # resume the most recent session for this dir
cool-code --resume <id>       # resume a specific saved session
cool-code --effort high       # set reasoning effort for this run
```

### Subcommands

```bash
cool-code scan                       # summarize project structure
cool-code scan --refresh --json      # refresh cache, JSON output
cool-code task "Add user auth"       # one-shot structured plan
cool-code config get llm.model
cool-code config set llm.model "gemini-2.5-flash"
cool-code config set llm.maxTokens 2048
cool-code skill install ./path/to/skill
cool-code skill install https://github.com/user/some-skill.git --global
```

### Agent modes

1. **Plan** - investigates with read-only tools and produces a detailed, file-level plan plus a task breakdown, without touching your code. When ready, offers **Start implementation** or **Keep planning**.
2. **Agent** (default) - executes tasks autonomously, applying edits and running commands.
3. **Ask** - read-only mode; mutating tools (edit, new file, shell, rename, replace) are blocked.

Switch mid-session with **Shift+Tab** (cycles Plan → Agent → Ask) or `/mode plan|agent|ask`.

### Interactive commands

Type `/` to open the command palette; press Tab to complete the highlighted command.

- `/help` - show available commands
- `/mode` - show or switch mode
- `/effort` - show or set reasoning effort (`minimal|low|medium|high|xhigh`)
- `/pin` / `/unpin` - pin a file's contents into context (or list/unpin)
- `/context` - preview context, pinned files, and token usage
- `/sessions` - list saved sessions for this directory
- `/install-skill` - install a skill from a local path or git URL (add `--global`)
- `/clear` - clear the screen
- `/exit` - exit

You can also type new messages while the agent is working - they're queued and picked up after the current step.

### Model providers

The provider is inferred from the model id when `llm.provider` is unset:

- `gpt-*`, `o1*`, `o3*` → OpenAI
- `claude-*` → Anthropic
- everything else (e.g. `gemini-*`) → Google

```bash
cool-code config set llm.model "gpt-4o"           # inferred as OpenAI
cool-code config set llm.model "claude-sonnet-5"  # inferred as Anthropic
cool-code config set llm.provider "openai"        # or set it explicitly
```

If the required key is missing, the CLI prints provider-specific setup guidance and exits.

### Project memory & skills

Create a `COOLCODE.md` in your project root for persistent, project-specific instructions (loaded every turn; global fallback `~/.coolcode/COOLCODE.md`).

Skills are reusable instruction modules under `.coolcode/skills/<name>/SKILL.md` (project) or `~/.coolcode/skills/<name>/SKILL.md` (global), each with optional `name`/`description` frontmatter. The agent sees a catalog and calls the `use_skill` tool to load one on demand. Install skills from a local path or git URL with `cool-code skill install` or `/install-skill`.

## Configuration (`.coolcode.json`)

```json
{
  "llm": {
    "temperature": 0.2,
    "maxTokens": 2048
  },
  "features": {
    "scanCache": true,
    "fileTreeMaxDepth": 4,
    "maxContextTokens": 20000
  }
}
```

Provider identity and security-sensitive settings (`llm.model`, `llm.provider`, `llm.baseUrl`, `llm.apiKeyEnv`, `llm.allowInsecureHttp`, `features.allowDangerous`, `features.confirmEdits`) are global-only. Set them with `cool-code config set`; repository-controlled values are ignored, and the CLI lists any it ignored at startup.

`guardrails.blockReadPatterns` is the one exception: a project file may **add** patterns, since that can only narrow what the agent may read. It can never remove one.

## Safety

- **Canonical path jail** - reads, searches, writes, pins, and added directories are restricted to explicitly granted roots after resolving symlinks.
- **Read guardrails** - `blockReadPatterns` applies to reads, searches, pins, context trees, and subagents. `.gitignore` is respected by trees and searches.
- **Trusted endpoints** - repositories cannot select proxy hosts, credential variables, guardrails, or bypass flags. Remote proxies require HTTPS and cross-origin redirects are disabled.
- **Command isolation** - every arbitrary shell command and project-code execution requires confirmation by default. Child processes receive a small environment allowlist without API keys, tokens, cookies, or cloud credentials.
- **Network isolation** - web fetches require HTTPS and reject loopback, private, link-local, multicast, and cloud metadata addresses, including after redirects and DNS resolution.
- **Data-loss prevention** - common credentials and sensitive environment values are redacted before provider egress, terminal rendering, and session persistence. Provider errors never include endpoint URLs or raw bodies.
- **Private persistence** - credentials and sessions use private directories and mode 0600 files; symlink-backed config, credential, session, memory, skill, and `.env` files are rejected.
- **Read-only modes** - Plan and Ask modes deterministically block mutating tools and project-code execution.

## Architecture

```
main.go                      # entry → cmd.Execute()
cmd/                         # cobra CLI: root + config / scan / skill / task
internal/
  llm/        provider-agnostic tool-calling over Anthropic / OpenAI / Gemini
  agent/      processor (agentic loop), context, prompts, danger gating, task plan
  tools/      24 tools + JSON-schema defs, path safety, registry
  config/     .coolcode.json load/save + dotted get/set
  session/    ~/.coolcode/sessions persistence
  skills/     SKILL.md discovery + install (local path / git URL)
  memory/     COOLCODE.md loader
  project/    gitignore matching, folder tree, project scan
  tui/        Bubble Tea app, Lipgloss theme, Glamour markdown, commands
  types/      shared domain types
```

The agent loop sends each provider the tool schemas and a running message list, executes the tool calls it returns, feeds results back, and repeats until the model produces a final message.

## Development

```bash
make build   # build ./cool-code
make test    # go test ./...
make vet     # go vet ./...
make fmt     # gofmt -w .
```

Requires Go 1.25+. External tools used at runtime when present: `git`, `rg` (ripgrep, for `find_symbol`), `bash`, and `npx prettier` (for `format_file`).

## Future scope

- **Semantic codebase index** - a local embedding index kept in sync via a Merkle tree of file hashes, exposed as a `codebase_search` tool.
- **Subscription sign-in for `/connect`** (Claude Pro/Max, ChatGPT/Codex OAuth).
- **Full-toolset subagents** (explore-only subagents shipped; write-capable ones need permission routing).
- **Edit checkpoints and `/undo`**, **granular permission allowlists**, **custom slash commands**, **MCP support**, **hooks**, **`@file` mentions**, **cost tracking**, and **multimodal input**.

## Changelog

### 2.2.1: Security hardening (2026-08)

- Trusted global-only proxy endpoints, credential binding, HTTPS validation, redirect blocking, and sanitized provider errors.
- Symlink-aware workspace jails across reads, searches, writes, pins, skills, memory, and multi-root access.
- Confirmation for arbitrary commands and project-code execution, with sensitive environment variables removed from child processes.
- SSRF protection for web fetches, including private networks, cloud metadata, redirects, and DNS rebinding checks.
- Secret redaction before model egress and session persistence; sessions and credentials now use private filesystem permissions.
- Deterministic read-only Plan and Ask modes, plus final-answer-only transcript rendering that suppresses intermediate model prose.

### 2.2 - Responsive TUI, proxy endpoints & multi-root workspaces (2026-08)

- Responsive compact TUI, persistent composer, bounded menus and overlays.
- Distinct `PLAN READY` cards and improved Markdown formatting.
- `/add-dir` multi-root workspace support and cross-directory completion.
- OpenAI-compatible proxy/base URL and API-key environment configuration.
- Reasoning effort via `--effort`, `/effort`, and `llm.reasoningEffort`.
- Terminal OSC reply filtering and render regression tests.

### 2.1 - Subagents, concurrency & quality of life (2026-07)

- **Explore subagents** - new `spawn_agent` tool; the agent fans out concurrent read-only mini-agents (own history, 15-iteration cap, cannot nest) with live per-agent status lines in the TUI.
- **Parallel tools** - independent read-only tool calls in one assistant turn run concurrently; mutating calls stay sequential with results in original call order.
- **Cancellation** - Esc or Ctrl+C aborts an in-flight turn (LLM request, shell commands, and web requests all stop); Ctrl+C while idle quits. Interrupted tool calls are closed cleanly so the next request never fails on unpaired tool calls.
- **Streaming** - assistant output streams token-by-token on all three providers (Anthropic, OpenAI, Gemini) over SSE.
- **`/connect`** - interactive provider setup; API keys stored in `~/.coolcode/credentials.json` (0600), provider/model defaults in `~/.coolcode/settings.json`, env vars as fallback. Starting without a key opens the TUI with a hint instead of exiting.
- **Reliability** - automatic retry with exponential backoff on 429/5xx/network errors (honors `Retry-After`); real token usage from provider responses in the status bar; sessions persist on quit and cancel; fixed a data race between the status bar and the agent loop.
- **TUI** - multiline input (Enter submits, Alt+Enter/Ctrl+J newline), mouse-wheel and PgUp/PgDn scrolling, input-history recall on Up/Down, `--resume` repopulates the visible transcript, markdown re-wraps on window resize, edit confirmations show red/green diff hunks.
- **Safety** - `edit_file` and `shell_command` directories are confined to the project root; `confirmEdits` works independently of `allowDangerous`; all tool error labels are now descriptive.

### 2.0 - Go rewrite

- Full rewrite from TypeScript/Node to a single Go binary: native provider function-calling (Anthropic, OpenAI, Gemini), Bubble Tea/Lipgloss/Glamour TUI, cobra CLI, plan/agent/ask modes, sessions, skills, project memory, and 23 tools. The TypeScript version lives on the `node` branch.
