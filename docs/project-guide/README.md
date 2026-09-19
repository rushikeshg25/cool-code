# Cool-Code Project Guide

> Generated: 2026-09-20 from commit `ed132aa` (103 tracked files before this guide).

## What this is

Cool-Code is a Go command-line coding agent for developers who want to ask questions, plan changes, or edit a local project through model-selected tools ([CLI entry](../../cmd/root.go#L39), [modes](../../internal/types/types.go#L7)). It runs either an interactive Bubble Tea terminal UI or a single non-interactive turn, sharing the same agent processor ([interactive](../../cmd/root.go#L109), [print](../../cmd/print.go#L58)). The processor repeatedly exchanges messages and structured tool calls with an Anthropic, OpenAI, or Google backend, executes permitted tools locally, and returns the answer ([loop](../../internal/agent/processor.go#L146), [provider selection](../../internal/llm/llm.go#L129)).

## The five-file tour

Read these five source files in order; follow their calls into the other packages as needed.

| # | File | Why this one | Then look at |
| --- | --- | --- | --- |
| 1 | [cmd/root.go](../../cmd/root.go#L39) | `main.go` immediately calls this CLI dispatcher; it assembles configuration, processor, session restoration, and UI. | [Startup](02-flow.md#startup) |
| 2 | [internal/tui/update.go](../../internal/tui/update.go#L496) | Shows how an entered prompt becomes a cancellable processor turn and how completion returns to the UI. | [Flow](02-flow.md) |
| 3 | [internal/agent/processor.go](../../internal/agent/processor.go#L146) | Owns the model/tool loop, execution policy, concurrency, and final answer. | [Boundaries](01-architecture.md#boundaries-and-contracts) |
| 4 | [internal/tools/registry.go](../../internal/tools/registry.go#L15) | Shows the callable tool surface, dispatch, and approval reasons; follow individual tool definitions from here. | [Tool files](03-structure.md#internal-tools) |
| 5 | [internal/session/session.go](../../internal/session/session.go#L16) | Defines what survives a restart and how session files are read and written. | [State and persistence](01-architecture.md#state-and-persistence) |

## Run it

Use Go 1.25.13 or a compatible newer toolchain, as declared in [go.mod:3](../../go.mod#L3). From the repository root:

```bash
go mod download
make build
./cool-code
# In the UI, use /connect to select a provider and enter its API key.

# After configuring a provider, run a read-only scripted query:
./cool-code --mode ask --print "Explain the agent loop"

# Development checks:
make test
make vet
```

Build and check targets come from [Makefile:9](../../Makefile#L9); `--print` and `--mode` are wired in [cmd/root.go:66](../../cmd/root.go#L66). The provider menu is defined in [internal/tui/app.go:140](../../internal/tui/app.go#L140). Interactive startup tolerates a missing key and offers `/connect`; print mode needs credentials immediately ([cmd/root.go:109](../../cmd/root.go#L109), [cmd/print.go:58](../../cmd/print.go#L58)).

Alternatively, configure the matching provider environment variable: `GOOGLE_GENERATIVE_AI_API_KEY`, `OPENAI_API_KEY`, or `ANTHROPIC_API_KEY` ([provider map](../../internal/llm/llm.go#L89)). Select non-default models with `./cool-code config set llm.model MODEL_ID`; model/provider/endpoint settings are written to trusted global settings ([scope rules](../../internal/config/config.go#L202)). See [credential precedence](01-architecture.md#boundaries-and-contracts) before using a proxy.

The core executable uses Go dependencies; individual tools additionally require programs such as Git, Bash, ripgrep, or a project-specific formatter/test runner. See the [tooling inventory](04-tech-stack.md). These commands are instructions for a maintainer, not a claim that a live provider session or the Go suite was run while preparing this guide.

## Reading order for this guide

1. [Architecture](01-architecture.md) — components, contracts, data, and runtime topology.
2. [Flow](02-flow.md) — startup, turns, approvals, subagents, persistence, and other commands.
3. [Structure](03-structure.md) — source files, exports, callers, and test coverage locations.
4. [Tech stack](04-tech-stack.md) — versions, external dependencies, build, CI, and releases.
5. [Decisions](05-decisions.md) — evidence-backed choices, tradeoffs, conventions, and gotchas.

## Open questions

- **Persistence failures:** should a failed session save be visible to the user? `persist` currently discards the error from `session.Save` ([app.go:429](../../internal/tui/app.go#L429)); the desired recovery behavior is not specified there.
- **Queued follow-ups:** should a prompt queued during a final-answer-only response automatically start a new turn? The queue is consumed after tool batches, while the final-answer branch exits before that point ([processor.go:191](../../internal/agent/processor.go#L191), [processor.go:359](../../internal/agent/processor.go#L359)). The current behavior is documented in the flow; intended behavior needs a product decision.
- **Concurrency budget:** what limit should apply to simultaneous read calls or explore agents? The parent starts a goroutine for each approved read/subagent call with no local worker-count limit; each child has a 15-iteration limit ([processor.go:251](../../internal/agent/processor.go#L251), [subagent.go:16](../../internal/agent/subagent.go#L16)). No throughput or cost target is specified by these implementations.
- **Documentation contract:** should the README's older statements about final-only output, memory loading, and context budgeting be revised? The implementation keeps intermediate assistant prose, loads memory when constructing context, and subtracts the estimated system size when choosing history ([processor.go:202](../../internal/agent/processor.go#L202), [context.go:40](../../internal/agent/context.go#L40), [context.go:190](../../internal/agent/context.go#L190)). See [Gotchas](05-decisions.md#gotchas).
- **Dependency and platform intent:** why is Lip Gloss pinned to its particular pseudo-version, and which external-command workflows are expected to work on Windows? The manifest supplies the pin but no rationale; the release matrix includes Windows while shell execution requires Bash ([go.mod:11](../../go.mod#L11), [release.yml:26](../../.github/workflows/release.yml#L26), [shell.go:30](../../internal/tools/shell.go#L30)).

This guide describes checked-in source. It does not establish provider availability, deployed release health, or performance under load.
