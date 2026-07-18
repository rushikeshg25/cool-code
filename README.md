# Cool-Code: a fast, native CLI coding agent

An intelligent command-line coding agent — describe what you want, and it reads, plans, and edits your codebase using live tools. Similar in spirit to Claude Code and Gemini CLI, now rewritten in **Go** with a **Bubble Tea** TUI and **native provider tool-calling**.

> **Rewritten in Go.** Cool-Code was previously a TypeScript/Node CLI. It is now a single statically-linked Go binary — no `node_modules`, faster start-up, real goroutine concurrency, and a polished Charm-stack terminal UI. The agent uses each provider's native function-calling API (Anthropic, OpenAI, Gemini) instead of parsing tool calls out of free text.

## Overview

Cool-Code combines large language models with a comprehensive set of development tools for an interactive development experience. It works with multiple model providers (Google, OpenAI, Anthropic) and finds context live through search and read tools — no vector database required.

## Features

- **Multiple model providers** — Google, OpenAI, or Anthropic. Connect one with `/connect` (keys stored in `~/.coolcode/credentials.json`) or export the matching env var; the provider is inferred from the model id.
- **Native tool-calling with streaming** — structured function-calling against each provider's API, with assistant output streamed token-by-token.
- **Explore subagents** — the agent can fan out several read-only `spawn_agent` explorers concurrently to investigate independent areas of a codebase, with live per-agent status in the TUI.
- **Concurrent tools + cancellation** — independent read-only tool calls run in parallel; Esc or Ctrl+C cancels a running turn without quitting.
- **Polished TUI** — a Bubble Tea terminal UI with Glamour-rendered markdown, multiline input (Alt+Enter), mouse-wheel scrolling, input-history recall, colored diff previews for edits, a `/` slash-command palette with Tab autocomplete, a live task panel, and Shift+Tab to switch modes.
- **Three agent modes** — Plan (read-only investigation → detailed plan), Agent (autonomous execution), Ask (read-only Q&A). After Plan mode produces a plan, choose **Start implementation** to jump straight into Agent mode.
- **Project memory (`COOLCODE.md`)** — persistent project instructions loaded into every prompt.
- **Skills** — discoverable, model-invoked instruction modules under `.coolcode/skills/` (compatible with Claude Code skills).
- **Web access** — `web_fetch` and `web_search` tools.
- **Session persistence** — conversations saved to `~/.coolcode/sessions/` (including on quit and cancel); resume with `--continue` / `--resume` and the prior conversation reappears in the transcript.
- **Reliability** — automatic retry with backoff on transient API errors, and real token usage from provider responses in the status bar.
- **Task tracking, input queuing, safety guardrails** — real-time checklists, mid-turn message queuing, and path/read/danger protections.

## Install

### Go install (recommended)

```bash
go install github.com/rushikeshg25/cool-code@latest
```

This places a `cool-code` binary in your `$GOBIN` (usually `~/go/bin` — make sure it's on your `PATH`).

### Build from source

```bash
git clone https://github.com/rushikeshg25/cool-code.git
cd cool-code
make build      # produces ./cool-code
# or: go build -o cool-code .
```

### Connect a provider

The easiest way: start `cool-code` and run **`/connect`** — pick a provider, paste your API key, and it's stored in `~/.coolcode/credentials.json` (mode 0600). The chosen provider and a default model are saved as global defaults in `~/.coolcode/settings.json`; a project `.coolcode.json` still wins when present.

Env vars keep working as a fallback:

```bash
export GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here   # Gemini (default)
# export OPENAI_API_KEY=your_api_key_here               # OpenAI
# export ANTHROPIC_API_KEY=your_api_key_here            # Anthropic
```

A local `.env` file in the project directory is loaded automatically (see `.env.example`).

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

1. **Plan** — investigates with read-only tools and produces a detailed, file-level plan plus a task breakdown, without touching your code. When ready, offers **Start implementation** or **Keep planning**.
2. **Agent** (default) — executes tasks autonomously, applying edits and running commands.
3. **Ask** — read-only mode; mutating tools (edit, new file, shell, rename, replace) are blocked.

Switch mid-session with **Shift+Tab** (cycles Plan → Agent → Ask) or `/mode plan|agent|ask`.

### Interactive commands

Type `/` to open the command palette; press Tab to complete the highlighted command.

- `/help` — show available commands
- `/mode` — show or switch mode
- `/pin` / `/unpin` — pin a file's contents into context (or list/unpin)
- `/context` — preview context, pinned files, and token usage
- `/sessions` — list saved sessions for this directory
- `/install-skill` — install a skill from a local path or git URL (add `--global`)
- `/clear` — clear the screen
- `/exit` — exit

You can also type new messages while the agent is working — they're queued and picked up after the current step.

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
    "model": "gemini-2.5-flash",
    "provider": "google",
    "temperature": 0.2,
    "maxTokens": 2048
  },
  "features": {
    "scanCache": true,
    "fileTreeMaxDepth": 4,
    "allowDangerous": false,
    "confirmEdits": false,
    "maxContextTokens": 20000
  },
  "guardrails": {
    "blockReadPatterns": [".env", ".env.*", "*.pem", "*.key", "id_rsa", "id_ed25519", ".npmrc"]
  }
}
```

`provider` is optional; when omitted it is inferred from `model`. `COOLCODE_DEBUG=1` enables non-fatal debug logs.

## Safety

- **Path validation** — writes are restricted to the project root; traversal escapes are rejected.
- **Read guardrails** — `blockReadPatterns` prevents reading sensitive files (including when pinned).
- **Gitignore aware** — respects `.gitignore` in the file tree and search tools.
- **Danger prompts** — risky shell commands (`rm -rf`, `sudo`, `git push --force`, `curl | bash`, …), non-dry-run bulk replaces, and overwriting renames require confirmation unless `--allow-dangerous`.
- **Ask mode** — strict read-only enforcement.

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

Requires Go 1.24+. External tools used at runtime when present: `git`, `rg` (ripgrep, for `find_symbol`), `bash`, and `npx prettier` (for `format_file`).

## Future scope

- **Semantic codebase index** — a local embedding index kept in sync via a Merkle tree of file hashes, exposed as a `codebase_search` tool.
- **Subscription sign-in for `/connect`** (Claude Pro/Max, ChatGPT/Codex OAuth).
- **Full-toolset subagents** (explore-only subagents shipped; write-capable ones need permission routing).
- **Edit checkpoints and `/undo`**, **granular permission allowlists**, **custom slash commands**, **MCP support**, **hooks**, **`@file` mentions**, **cost tracking**, and **multimodal input**.
