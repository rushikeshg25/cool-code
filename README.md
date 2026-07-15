# Cool-Code: CLI Coding AI Agent

An intelligent command-line interface that leverages AI to help you interact with codebases and work on them, similar to Gemini CLI and Claude Code.

## Overview

Cool-Code combines large language models with a comprehensive set of development tools to provide an interactive development experience. Describe what you want to accomplish, and the agent understands your intent and executes the necessary operations. It works with multiple model providers (Google, OpenAI, Anthropic) and finds context live through search and read tools, so no vector database is required.

## Features

- **Multiple Model Providers**: Use Google, OpenAI, or Anthropic models. Set the model and provide the matching API key.
- **Interactive TUI**: An Ink-based terminal UI with `/` slash commands, an autocomplete dropdown, and Shift+Tab to switch modes.
- **Agent Modes**: Switch between Plan, Agent, and Ask modes. After Plan mode produces a plan, pick "Start implementation" to jump straight into Agent mode.
- **Reasoning Effort**: Toggle `low` / `high` effort with **Ctrl+E** (or `/effort`). High effort asks the model to investigate more thoroughly and self-verify; low biases to fast, direct action. Effort is independent of the mode and is remembered across sessions.
- **Project Memory (`COOLCODE.md`)**: Persistent project instructions automatically loaded into context.
- **Skills**: Discoverable, model-invoked instruction modules under `.coolcode/skills/`.
- **Web Access**: `web_fetch` and `web_search` tools for reading pages and searching the web.
- **Session Persistence**: Conversations are saved and can be resumed with `--continue` / `--resume`.
- **Task Tracking**: Real-time checklists for complex, multi-step operations.
- **Input Queuing**: Send new instructions while the agent is still processing a previous request.
- **Intelligent Code Analysis**: Understands your project structure and coding patterns without a vector database.
- **File Operations**: Read, create, edit, rename, and search files with AI assistance.
- **Shell Command Execution**: Run system commands safely through the agent.
- **Context-Aware**: Maintains conversation history and project context automatically.

## Setup Guide

### Prerequisites

- Node.js 18+
- An API key for at least one supported provider:
  - Google Gemini: `GOOGLE_GENERATIVE_AI_API_KEY`
  - OpenAI: `OPENAI_API_KEY`
  - Anthropic: `ANTHROPIC_API_KEY`

### Quick Install (Recommended)

Install globally from npm:

```bash
npm install -g cool-code
```

Set the API key for your chosen provider, for example:

```bash
export GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here
# or
export OPENAI_API_KEY=your_api_key_here
# or
export ANTHROPIC_API_KEY=your_api_key_here
```

### Development Setup

1. Clone the repository:

```bash
git clone https://github.com/rushikeshg25/cool-code.git
cd cool-code
```

2. Install dependencies:

```bash
npm install
```

3. Set up environment variables:

```bash
cp .env.example .env
```

Edit `.env` and add your provider API key, for example:

```
GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here
```

4. Build the project:

```bash
npm run build
```

5. Link for local development:

```bash
npm link
```

6. Run the tests:

```bash
npm test
```

## Model Providers

Cool-Code resolves the provider from `llm.provider` in `.coolcode.json`, or infers it from the model id when `provider` is not set:

- `gpt-*`, `o1*`, `o3*` -> OpenAI
- `claude-*` -> Anthropic
- everything else (e.g. `gemini-*`) -> Google

To switch providers, set the model and export the matching API key:

```bash
cool-code config set llm.model "gpt-4o"        # inferred as OpenAI
export OPENAI_API_KEY=your_api_key_here

cool-code config set llm.model "claude-sonnet-4-6"   # inferred as Anthropic
export ANTHROPIC_API_KEY=your_api_key_here
```

You can also set the provider explicitly:

```bash
cool-code config set llm.provider "openai"
```

If the required key is missing, the CLI prints provider-specific setup guidance.

## Usage

### Starting the CLI

Navigate to your project's directory and run:

```bash
cool-code
```

On startup, the CLI asks to initialize in the current directory, then lets you choose an operational mode before building context.

### Command Line Options

```bash
# Auto-initialize in current dir
cool-code --yes

# Exit without initializing
cool-code --no-init

# Reduce UI output for automation
cool-code --quiet

# Allow dangerous actions without prompts
cool-code --allow-dangerous

# Copy final response to clipboard
cool-code --copy

# Resume the most recent session for this directory
cool-code --continue

# Resume a specific saved session
cool-code --resume <session-id>
```

### Agent Modes

Cool-Code supports three distinct modes of operation:

1. **Plan Mode**: Investigates the codebase with read-only tools and produces a detailed, file-level plan (with a per-task breakdown) without touching your code. When the plan is ready, it offers next-step actions: **Start implementation** (switches to Agent mode and proceeds) or **Keep planning** (stay in Plan mode to refine).
2. **Agent Mode**: The default mode. Executes tasks autonomously, applying code changes and running commands.
3. **Ask Mode**: Read-only mode for questions and explanations. Mutating tools (edit, new file, shell, rename, replace) are blocked.

The mode chosen at startup is applied to the system prompt. Switch mid-session with **Shift+Tab** (cycles Plan -> Agent -> Ask) or the `/mode` command:

- `/mode plan`
- `/mode agent`
- `/mode ask`

### Reasoning Effort

Effort is a separate axis from the mode — you can run, for example, a high-effort Agent session. Toggle it with **Ctrl+E** or set it explicitly with `/effort low|high`. At **high** effort the agent is instructed to read the relevant files before acting, chain read-only search tools, and verify changes with the project's tests/build where feasible; **low** keeps things fast and terse. The active mode and effort are shown in the status bar and persisted with the session.

### Project Memory (COOLCODE.md)

Create a `COOLCODE.md` file in your project root to give the agent persistent, project-specific instructions (conventions, do/don't rules, architecture notes). It is loaded into the system prompt on every turn. A global `~/.coolcode/COOLCODE.md` is used as a fallback when no project file exists.

### Skills

Skills are reusable instruction modules the model can load on demand. Create them under:

```
.coolcode/skills/<name>/SKILL.md      # project-level
~/.coolcode/skills/<name>/SKILL.md    # global
```

Each `SKILL.md` may start with simple frontmatter:

```markdown
---
name: pdf-filler
description: Fill PDF forms from a data file
---

Step-by-step instructions the agent should follow when this skill applies...
```

The agent sees a catalog of available skill names and descriptions, and calls the `use_skill` tool to load a skill's full instructions into context when it is relevant.

**Installing skills.** You can install skills (including ones authored for Claude Code) from a local path or a git URL. A source may be a single `SKILL.md`, a skill directory, or a repository containing several skills.

```bash
# From the CLI
cool-code skill install ./path/to/skill
cool-code skill install https://github.com/user/some-skill.git
cool-code skill install ./skill --global    # install to ~/.coolcode/skills

# Or from inside the interactive session
/install-skill ./path/to/skill
/install-skill https://github.com/user/some-skill.git --global
```

Installed skills are copied into `.coolcode/skills/` (or `~/.coolcode/skills/` with `--global`) and become available immediately.

### Web Access

The agent can reach the web through two tools:

- `web_fetch`: retrieves an `http(s)` URL and returns it as readable text (HTML is stripped).
- `web_search`: searches the web and returns result titles, URLs, and snippets.

### Sessions

Conversations (messages, summary, pinned files, mode, and effort) are saved automatically to `~/.coolcode/sessions/<id>.json` (written owner-only, `0600`) after each turn. Resume them with `--continue` (most recent for the current directory) or `--resume <id>`, and list saved sessions inside the CLI with `/sessions`.

### Advanced CLI Commands

```bash
# Project summary and framework detection
cool-code scan

# Force refresh the scan cache
cool-code scan --refresh

# JSON output
cool-code scan --json

# Task planning
cool-code task "Add user authentication and password reset"

# Config management
cool-code config get llm.model
cool-code config set llm.model "gemini-2.5-flash"
cool-code config set llm.maxTokens 2048

# Install a skill (local path or git URL)
cool-code skill install ./path/to/skill
cool-code skill install https://github.com/user/some-skill.git --global
```

### Interactive Commands

Inside the CLI prompt, type `/` to open an autocomplete dropdown of commands; press Tab to complete the highlighted one.

- `/help` Show available commands
- `/mode` Show or switch the current mode (`/mode plan|agent|ask`)
- `/effort` Show or switch reasoning effort (`/effort low|high`)
- `/pin` Pin a file's contents into context (e.g. `/pin src/index.ts`)
- `/unpin` Unpin a file (or list pinned files)
- `/context` Preview the prompt context, pinned files, and token usage
- `/sessions` List saved sessions for this directory
- `/install-skill` Install a skill from a local path or git URL (add `--global`)
- `/clear` Clear the screen
- `/exit` Exit

Press **Shift+Tab** at any time to cycle the mode (Plan -> Agent -> Ask), or **Ctrl+E** to toggle effort (low -> high).

### Non-blocking Input

You can type and send new messages even while the agent is processing tool calls. These messages are queued and processed immediately after the current step completes, allowing you to pivot or provide feedback mid-task.

## Architecture

### Core Components

#### 1. Entry Point (`src/index.ts`)
Initializes the CLI using Commander.js, wires up flags (including `--continue` / `--resume`), and starts the interactive session.

#### 2. UI Layer (`src/ui/`)
- **App (`app.tsx`)**: The Ink (React) terminal UI — header, output log, input with the `/` autocomplete dropdown, Shift+Tab mode cycling, command dispatch, confirmations, and the post-plan action menu.
- **Commands (`commands.ts`)**: Single source of truth for slash commands (used by the dropdown, `/help`, and the dispatcher).
- **Landing (`landing.ts`)**: Compact welcome screen.
- **Spinner (`spinner.ts`)**: The `StatusReporter` interface plus a console spinner used for CLI/quiet output.
- **Theme (`theme.ts`)**: Single-accent palette and glyph constants shared across the console and Ink surfaces.
- **Clipboard / Markdown (`clipboard.ts`, `utils/markdown.ts`)**: Output helpers.

The Ink/React stack is kept external in the bundle and loaded from `node_modules` at runtime (it cannot be esbuild-bundled), and is lazy-loaded so non-interactive CLI commands stay light.

#### 3. Core Engine (`src/core/`)
- **Processor (`processor.ts`)**: Orchestrates query turns, tool execution, mode handling, and message queuing.
- **LLM (`llm.ts`)**: Resolves the model provider (Google / OpenAI / Anthropic) and streams responses.
- **Context Manager (`contextManager.ts`)**: Builds the prompt (system, tools, skills, project state, conversation), manages the token budget, summarization, and pinned files.
- **Project Memory (`projectMemory.ts`)**: Loads `COOLCODE.md`.
- **Skills (`skills.ts`, `skillInstaller.ts`)**: Discovers/parses `SKILL.md` files and installs skills from local paths or git URLs.
- **Session (`session.ts`)**: Persists and restores conversations.
- **Prompts (`prompts.ts`)**: System prompt, mode definitions, and tool instructions.

#### 4. Tool System (`src/core/tools/`)
- **File Operations**: `readFileTool`, `editTool`, `newFileTool`, `renameFileTool`, `replaceInFilesTool`, `openFileAtTool`.
- **Search & Discovery**: `globTool`, `grepTool`, `findSymbolTool`, `listRecentFilesTool`.
- **Git**: `gitStatusTool`, `gitDiffTool`, `gitCommitTool`.
- **Code Quality**: `formatFileTool`, `lintFixTool`, `runTestsTool`.
- **Project**: `projectSummaryTool`, `newModuleTool`, `addScriptTool`, `generateReadmeSectionTool`.
- **Web**: `webFetchTool`, `webSearchTool`.
- **Skills**: `useSkillTool`.
- **System**: `shellTool`, `ignoreGitIgnoreFileTool`.
- **Infrastructure**: `tool-registery.ts`, `toolValidator.ts`, `toolUtils.ts`.

## Configuration

### Environment Variables
- `GOOGLE_GENERATIVE_AI_API_KEY`, `OPENAI_API_KEY`, or `ANTHROPIC_API_KEY`: API key for your chosen provider.
- `COOLCODE_DEBUG=1`: Print non-fatal debug logs (otherwise suppressed to keep the UI clean).

### Project Config (.coolcode.json)

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
    "blockReadPatterns": [
      ".env",
      ".env.*",
      "*.pem",
      "*.key",
      "id_rsa",
      "id_ed25519",
      ".npmrc"
    ]
  }
}
```

`provider` is optional; when omitted it is inferred from `model`.

## Safety Features

- **Path Confinement**: Read, edit, and search tools (`read_file`, `edit_file`, `grep`, `find_symbol`, and the write tools) are restricted to the project root. The confinement check resolves symlinks, so a symlink inside the root cannot be used to escape it.
- **Read Guardrails**: `blockReadPatterns` prevents reading sensitive files (e.g. `.env`, `*.key`), including when pinned.
- **SSRF Protection**: `web_fetch` blocks requests to loopback, private, link-local, and cloud-metadata addresses (`169.254.169.254`), and re-validates every redirect hop.
- **Restrictive File Permissions**: Session, config, and scan-cache files are written owner-only (`0600`).
- **Hardened Skill Install**: Only `https`/`git@`/`ssh` git sources are accepted; command-helper transports (`ext::`, `file://`) are rejected.
- **Prototype-Pollution Guard**: `config set` rejects `__proto__`/`constructor`/`prototype` key paths.
- **Git Integration**: Automatically respects `.gitignore` patterns.
- **Dangerous Action Prompts**: Requires explicit confirmation for potentially destructive shell commands.
- **Edit Confirmations**: Configurable prompts for code edits.
- **Shell Argument Escaping**: Tool-built shell commands single-quote interpolated values.
- **Ask Mode**: Strict read-only enforcement for safer exploration.

## Future Scope

Planned and candidate features for future iterations:

- **Semantic codebase index (Cursor-style)**: Build a local embedding index of the codebase for natural-language code search, and keep it in sync incrementally using a **Merkle tree of file content hashes** so only changed files are re-embedded. Embeddings would reuse the configured model provider's API key, exposed to the agent as a `codebase_search` tool. This is the main planned upgrade to retrieval.
- **Edit checkpoints and undo**: Snapshot files before mutating tools and allow `/undo` to roll back the agent's changes.
- **Granular permission allowlist**: Rule-based allow/deny per tool and per shell command, extending the current danger prompts.
- **Custom slash commands**: User-defined command templates under `.coolcode/commands/*.md`.
- **MCP (Model Context Protocol) support**: Connect external tool servers.
- **Subagents**: Spawn focused sub-agents for isolated sub-tasks.
- **Go rewrite for concurrency**: Rewrite the core in Go to exploit real concurrency, primarily to spin up subagents that run in parallel.
- **Hooks**: Run user-defined commands on events (e.g. before/after a tool runs).
- **Real streaming output**: Stream the model's final text to the terminal as it is generated.
- **`@file` mentions**: Inline file references in a query that auto-inject file contents.
- **Usage and cost tracking**: Per-turn token usage and estimated cost.
- **Multimodal input**: Accept images/screenshots as part of a query.

## Project Structure

```
cool-code/
├── src/
│   ├── core/                 # Core engine
│   │   ├── tools/            # Tool implementations
│   │   ├── utils/            # Utility functions
│   │   ├── contextManager.ts
│   │   ├── llm.ts
│   │   ├── processor.ts
│   │   ├── projectMemory.ts
│   │   ├── skills.ts
│   │   ├── skillInstaller.ts
│   │   ├── session.ts
│   │   └── prompts.ts
│   ├── types/                # TypeScript definitions
│   ├── ui/                   # UI components
│   └── index.ts              # Entry point
├── dist/                     # Compiled output
├── package.json
└── README.md
```

## Testing

The project uses [Vitest](https://vitest.dev/):

```bash
npm test           # run the test suite
npm run type-check # TypeScript type checking
npm run build      # bundle with tsup
```
