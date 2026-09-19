# Decisions

These choices describe the inspected implementation. Explicit comments support some rationales; the remaining rationales are marked as inferences.

## Own the provider protocol behind a small interface

- **What:** The agent uses common `Request`, `Message`, `ToolDef` and `Provider` types. `Streamer` is an optional capability returning the same assembled message as `Complete`; concrete adapters use shared HTTP/SSE helpers.
- **Evidence:** [`llm.go:41`](../../internal/llm/llm.go#L41), [`llm.go:68`](../../internal/llm/llm.go#L68), [`http.go:87`](../../internal/llm/http.go#L87), [`http.go:169`](../../internal/llm/http.go#L169), [processor capability check](../../internal/agent/processor.go#L172).
- **Why, apparently:** Inferred: keep the agent loop independent of provider wire formats while permitting a simple fake provider in tests.
- **Tradeoff:** Small shared API and no provider SDK dependency, but changes to provider protocols and streaming formats must be maintained here.
- **Confidence:** Interface and fake-provider seam confirmed in code; rationale inferred. [Test fake](../../internal/agent/processor_test.go#L21).

## Treat repository configuration as less trusted than user settings

- **What:** Load defaults and global settings, then merge only safe project preferences. Project read patterns can add restrictions; project config cannot choose provider identity, endpoint, credential variable or confirmation bypasses. A custom endpoint cannot inherit a first-party stored key.
- **Evidence:** [Load contract](../../internal/config/config.go#L89), [project merge](../../internal/config/config.go#L292), [proxy credential boundary](../../internal/llm/llm.go#L167), [restricted `.env` loading](../../cmd/root.go#L169).
- **Why, apparently:** Explicit comments say opening an untrusted repository must not redirect credentials or disable guardrails.
- **Tradeoff:** A checked-in config cannot completely reproduce the user's provider setup. Set security-sensitive values through the config command/global settings, and supply proxy credentials explicitly.
- **Confidence:** Confirmed by comments and implementation.

## Enforce capabilities at execution, and use canonical paths for I/O

- **What:** Plan/Ask block mutating tools in the processor; read-only workers dispatch through `RunReadOnly`, which rejects unknown or unauthorized tools. File tools receive a resolved allowed path, rather than validating one spelling and opening another. `.git`, `.coolcode` and `.coolcode.json` are protected against tool writes.
- **Evidence:** [mode gate](../../internal/agent/processor.go#L231), [read-only dispatch](../../internal/tools/registry.go#L65), [canonical read contract](../../internal/tools/safety.go#L68), [protected writes](../../internal/tools/safety.go#L87).
- **Why, apparently:** Explicit comments distinguish actual execution permissions from tool definitions advertised to the model, and describe symlink/path alias escapes and persistent code execution through Git configuration.
- **Tradeoff:** Adding a tool requires updating its permission metadata and using the proper resolver; legitimate edits to protected operational files need an action outside the agent file tools.
- **Confidence:** Confirmed by comments and implementation. This describes application checks, not a separate OS sandbox around approved subprocesses.

## Separate approved shell text from assembled command arguments

- **What:** User-confirmed command strings use `bash -c`. Commands constructed from filenames or other model/repository values use direct argv. Child processes receive an environment allowlist, and resolved test/lint commands are fixed before approval.
- **Evidence:** [`shell.go:23`](../../internal/tools/shell.go#L23), [environment allowlist](../../internal/tools/shell.go#L100), [command preparation before gate](../../internal/agent/processor.go#L234), [complete escaped approval text](../../internal/tools/registry.go#L74).
- **Why, apparently:** Explicit comments identify command injection and hidden executable text as the threats. Formatter confirmation exists because `npx prettier` can resolve code supplied by the repository.
- **Tradeoff:** Shell pipelines remain usable when approved, but project commands requiring secrets from inherited environment variables may fail. The binary also needs Bash for these commands, even on a platform with a release artifact.
- **Confidence:** Confirmed by comments and implementation. [Formatter rationale](../../internal/tools/registry.go#L95).

## Parallelize reads, then execute mutations in order

- **What:** Each model response is partitioned after a sequential permission/approval pass. Read-only calls and subagents run concurrently; mutations run afterward in their original relative order. Results are inserted into conversation history in original call order.
- **Evidence:** [pre-pass](../../internal/agent/processor.go#L218), [read workers](../../internal/agent/processor.go#L251), [mutation loop](../../internal/agent/processor.go#L320), [ordered result insertion](../../internal/agent/processor.go#L339).
- **Why, apparently:** Explicit comments require one confirmation overlay at a time and deterministic call/result pairing. Reduced read latency without concurrent writes is an additional inferred benefit.
- **Tradeoff:** Execution order differs from the model's total call order: a read listed after a write in the same batch still runs first. A model must use a later round for a read that depends on a mutation. Each batch creates workers without an explicit concurrency semaphore.
- **Confidence:** Scheduling behavior confirmed; latency rationale inferred.

The sequential group also includes tools without an explicit read-only classification. For example, `use_skill` declares neither capability flag: it can run in the parent loop but is absent from explore agents' read-only catalog ([misc_tools.go:96](../../internal/tools/misc_tools.go#L96), [processor.go:243](../../internal/agent/processor.go#L243), [subagent.go:38](../../internal/agent/subagent.go#L38)).

## Limit subagents to bounded, read-only exploration

- **What:** Subagents share the provider and allowed workspace context but start their own message history. Their tool surface excludes `spawn_agent`, and runtime dispatch enforces read-only access. Each explorer has a 15-iteration limit.
- **Evidence:** [`subagent.go:16`](../../internal/agent/subagent.go#L16), [tool surface](../../internal/agent/subagent.go#L35), [isolated loop](../../internal/agent/subagent.go#L49), [runtime dispatch](../../internal/agent/subagent.go#L79).
- **Why, apparently:** Inferred: allow parallel investigation without recursive fan-out or routing write approvals from child agents.
- **Tradeoff:** Exploration has a bounded number of rounds, but several explorers can still increase provider traffic. They cannot implement changes, nest agents, or use the parent's full conversation as implicit context.
- **Confidence:** Limits confirmed in code; motivation inferred.

## Preserve message protocol integrity across cancellation and compaction

- **What:** Interrupted turns get synthetic results for outstanding tool calls. History windows and retained compacted history begin at a real user message. Compaction summarizes older text and aims to keep a suffix from the latest 20 messages.
- **Evidence:** [deferred closure](../../internal/agent/processor.go#L154), [history selection](../../internal/agent/context.go#L188), [compaction limits](../../internal/agent/context.go#L269), [summary application](../../internal/agent/context.go#L319).
- **Why, apparently:** Comments explicitly say providers reject dangling tool calls. Compaction comments identify both information loss and oversized summarization requests as concerns.
- **Tradeoff:** The summary is lossy; starting at a user boundary can retain fewer than 20 messages. Token budgeting is approximate (`bytes / 4`, rounded up), with minimum-history and latest-turn fallbacks that can exceed the target.
- **Confidence:** Confirmed by comments and implementation. [Token estimate](../../internal/agent/context.go#L57).

## Buffer streaming output at line boundaries for redaction

- **What:** Provider fragments accumulate until a newline; complete lines are redacted before being reported. The final partial line is flushed at completion.
- **Evidence:** [`processor.go:723`](../../internal/agent/processor.go#L723).
- **Why, apparently:** Explicit comments explain that a credential may be split across arbitrary provider fragments, so fragment-by-fragment redaction could miss it.
- **Tradeoff:** The user sees output line by line; a long line may stay invisible until completion. This does not turn redaction into a proof that every possible secret format is detected.
- **Confidence:** Confirmed by comments and implementation.

## Keep durable state in ordinary private files

- **What:** Sessions and credentials are JSON files under the user's `.coolcode` directory, with directories mode `0700` and files mode `0600`; symlink-backed locations are rejected. Session messages remain an opaque agent-owned JSON field.
- **Evidence:** [session schema](../../internal/session/session.go#L14), [session save](../../internal/session/session.go#L45), [credential save](../../internal/creds/creds.go#L59).
- **Why, apparently:** Inferred: keep installation and session inspection simple while avoiding a database service and separating persistence from message implementation.
- **Tradeoff:** Session save uses direct file writes, not a transactional database or coordinated multi-process store. A missing/corrupt session loads as `nil` rather than exposing a recovery path.
- **Confidence:** Storage behavior confirmed; simplicity rationale inferred. [Load behavior](../../internal/session/session.go#L76).

## Gotchas

- **Skill catalog and skill body have different trust framing.** The system prompt wraps the catalog in untrusted-content markers; `use_skill` returns the selected body with an instruction prefix and no such wrapper. The README's blanket statement about skills does not describe both paths. [Catalog](../../internal/agent/context.go#L148) · [body](../../internal/tools/misc_tools.go#L121) · [README](../../README.md#L217).
- **`--yes` currently has no startup effect.** The flag is registered, but interactive startup does not read it; `--no-init`, separately, exits early. [Registration](../../cmd/root.go#L59) · [startup](../../cmd/root.go#L80).
- **README streaming descriptions conflict with the current loop.** The feature list says only final answers reach the transcript and scratch prose is dropped, but the processor explicitly retains prose before tool calls. A regression test asserts that behavior. [README:14](../../README.md#L14) · [README:22](../../README.md#L22) · [processor:202](../../internal/agent/processor.go#L202) · [test:145](../../internal/agent/processor_test.go#L145).
- **Context budget is approximate and includes a system-prompt deduction.** Contrary to the README's explanation, `window()` subtracts the estimated system text from `maxContextTokens`, then enforces a minimum conversation budget and may fall back to the latest user turn. Do not use this setting as a strict provider context limit. [README:189](../../README.md#L189) · [context:193](../../internal/agent/context.go#L193).
- **Memory is included every request but loaded once per context-manager construction.** Editing `COOLCODE.md` during a running session does not automatically reload the cached text. The root README's “loaded every turn” wording is misleading. Project memory falls back to global memory; the two are not concatenated. [README:167](../../README.md#L167) · [context:52](../../internal/agent/context.go#L52) · [prompt inclusion](../../internal/agent/context.go#L140) · [memory fallback](../../internal/memory/memory.go#L15).
- **Missing credentials do not necessarily exit.** Interactive startup constructs the processor with `AllowMissingKey: true` and opens the TUI with a `/connect` notice. This differs from the root README's general missing-key statement. [README:163](../../README.md#L163) · [root:109](../../cmd/root.go#L109) · [root:121](../../cmd/root.go#L121).
- **Extra roots do not make primary-root search global.** The system context explicitly tells the model to use absolute paths for extra directories and says search tools cover only the primary root. [context:156](../../internal/agent/context.go#L156).
- **Release publication depends on README formatting.** The tag's version must have a nonempty `### <version>:` section; publication stops if the extracted notes are empty. [release:88](../../.github/workflows/release.yml#L88).
- **The web-fetch tool description is broader than its validation.** Its schema says `http(s)`, but execution calls the HTTPS-only URL validator. A loopback HTTP LLM proxy allowance does not grant the web tool the same access. [description](../../internal/tools/web_tools.go#L46) · [validation call](../../internal/tools/web_tools.go#L64) · [provider exception](../../internal/llm/llm.go#L186).

## Conventions

- Tests live beside implementation files and use Go's `testing` package, temporary directories and small local fakes. The processor tests script provider messages and capture reporter events to assert the complete tool loop. [Example](../../internal/agent/processor_test.go#L21).
- Provider protocol tests use `httptest` SSE servers and temporarily replace package endpoint variables. Restore any changed globals after each test; parallel tests sharing those globals would need isolation. [Example](../../internal/llm/stream_test.go#L39).
- Tool declarations combine name, schema, capability flag and implementation; the registry exposes a stable ordered catalog and name lookup. Add tools through that convention and review read-only classification separately from advertising a schema. [Registry](../../internal/tools/registry.go#L14) · [web tool declaration](../../internal/tools/web_tools.go#L46).
- Routine checks are formatting, vet, race tests, static security analysis, vulnerability checking, build and a version smoke test. CI documents the actual sequence; `make test` alone does not include the race detector. [CI](../../.github/workflows/ci.yml#L25) · [Makefile](../../Makefile#L15).

Back to the [guide index](README.md) · [Architecture](01-architecture.md) · [Source map](03-structure.md).
