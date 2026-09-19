# Structure

## What lives where

This is one Go module and one executable. [`main.go`](../../main.go#L6) enters the Cobra command layer; [`internal/agent`](#internal-agent) owns the model/tool loop, [`internal/tools`](#internal-tools) implements capabilities, and [`internal/tui`](#internal-tui) presents and persists interactive sessions. Provider protocols live behind [`internal/llm`](#internal-llm), while smaller packages own configuration, credentials, project inspection, instruction files, and shared types. Follow [the flow guide](02-flow.md) for execution order and [the architecture](01-architecture.md) for contracts.

```text
main.go              executable entry point
cmd/                 interactive/print startup and administrative subcommands
internal/
  agent/             conversation, context, tool scheduling, exploration
  config/            trusted global and restricted project settings
  creds/             provider credential file
  llm/               provider adapters, HTTP/SSE, usage metadata
  memory/            COOLCODE.md loader
  project/           tree, ignore matcher, lightweight scanner
  security/          secret redaction and terminal sanitization
  session/           JSON conversation snapshots
  skills/            SKILL.md discovery and installation
  tools/             filesystem, process, Git, search, web capabilities
  tui/               Bubble Tea model, commands, rendering, persistence wiring
  types/             mode, task, and tool-result contracts
.github/             CI, release, and dependency-update workflows
Makefile, go.mod      build/check commands and dependency versions
docs/project-guide/  this guide
```

The tables list every production Go file. “Package-private” means the file's key symbols are used inside its package rather than exported as a public Go API. Test rows identify the behavior under test; their entry points are `Test*` functions called by `go test`.

## root

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`main.go`](../../main.go#L6) | Hands process control to the CLI. | `main` (package-private executable entry) | Go runtime |
| [`go.mod`](../../go.mod#L1) | Module identity, Go version, direct and indirect dependency requirements. | Module manifest | Go build/test/module tooling |
| [`Makefile`](../../Makefile#L1) | Local build, install, formatting, tests, vet, security checks; injects CLI version. | `build`, `install`, `test`, `vet`, `sec`, `fmt`, `tidy`, `run`, `clean` targets | Developer; CI calls `sec` |
| [`README.md`](../../README.md#L1) | User-facing installation, commands, configuration, and release history. | Release notes headings consumed by release workflow | Users and release automation |
| [`SECURITY.md`](../../SECURITY.md#L1) | Reporting instructions, trust model, release-integrity description. | Documentation | Maintainers and reporters |

## cmd

Cobra handlers select the interactive TUI or one-turn print mode; administrative commands operate without the TUI.

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`root.go`](../../cmd/root.go#L39) | Registers flags/subcommands; loads permitted `.env` keys; constructs processor; selects and restores sessions, checking the working directory. | `Execute`, `Version`; private `runInteractive`, `loadEnv`, `startingMode`, `restoreConversation` | `main`; other command handlers reuse startup helpers |
| [`print.go`](../../cmd/print.go#L26) | Reads argument/stdin prompt, runs one query, prints text/JSON and optional stderr progress. | Package-private `runPrint`, `printReporter`, `parseMode`, `writePlainText` | Root command; startup mode parser; task output |
| [`config.go`](../../cmd/config.go#L13) | Implements dotted-path configuration get/set. | Package-private `configCmd` | `Execute` |
| [`scan.go`](../../cmd/scan.go#L17) | Scans project or uses cached JSON; safely replaces cache via `os.Root`. | Package-private `scanCmd`, `readScanCache`, `writeScanCache` | `Execute` |
| [`skill.go`](../../cmd/skill.go#L14) | Installs local/Git skills into project or global skill directories. | Package-private `skillCmd` | `Execute` |
| [`task.go`](../../cmd/task.go#L18) | Requests a standalone structured task plan and prints text/JSON. | Package-private `taskCmd`, `printTaskPlan` | `Execute` |
| [`root_test.go`](../../cmd/root_test.go) | Checks selective environment import and same-directory symlink comparison. | `Test*` | `go test` |
| [`print_test.go`](../../cmd/print_test.go) | Checks prompt source precedence, modes, and stdout/stderr separation. | `Test*` | `go test` |
| [`mode_security_test.go`](../../cmd/mode_security_test.go) | Checks explicit startup/resume mode authority and invalid-mode rejection. | `Test*` | `go test` |
| [`scan_security_test.go`](../../cmd/scan_security_test.go) | Checks normal cache round-trip and refusal to clobber linked targets. | `Test*` | `go test` |
| [`terminal_security_test.go`](../../cmd/terminal_security_test.go) | Checks sanitization of plain command output. | `Test*` | `go test` |

## internal-agent

The processor coordinates capabilities through provider and reporter interfaces. Context assembly and conversation state remain package-private.

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`processor.go`](../../internal/agent/processor.go#L23) | Builds provider/tool surface; loops over completions; gates and schedules tools; tracks usage/tasks; handles compaction, queued input, roots, snapshots, and line-buffered streaming. | `Reporter`, `Options`, `Processor`, `Status`, `New`; processor methods including `ProcessQuery`, `Compact`, `Snapshot`, `Restore`, `AddDir`, `ConfigureLLM` | `cmd`, `tui` |
| [`context.go`](../../internal/agent/context.go#L21) | Loads tree, instructions and skills; builds system prompt; windows messages; handles summaries, pins, and extra directory context. | Package-private `contextManager`, `newContextManager` | `Processor` and subagent setup |
| [`subagent.go`](../../internal/agent/subagent.go#L14) | Defines `spawn_agent` and a bounded, non-nesting exploration loop with read-only dispatch. | Package-private `spawnAgentDef`, `runSubagent`, `parseSpawnTask` | Processor tool-definition construction and spawn scheduling |
| [`prompts.go`](../../internal/agent/prompts.go#L5) | Base safety/workflow prompt, mode instructions, and explore-subagent prompt. | Package-private `basePrompt`, `modePrompts`, `subagentPrompt` | Context manager; subagent loop |
| [`messages.go`](../../internal/agent/messages.go#L10) | Human-readable waiting and tool activity labels. | Package-private `waitingLabel`, `toolStatus` | Processor progress reporting |
| [`taskplan.go`](../../internal/agent/taskplan.go#L13) | One completion requesting JSON task-plan data; tolerates surrounding fences/preamble when parsing. | `TaskPlanStep`, `TaskPlan`, `CreateTaskPlan` | `cmd/task.go` |
| [`agent_test.go`](../../internal/agent/agent_test.go) | Covers parsing, context windows, prompt trust markers, streaming redaction and summary budgeting. | `Test*` | `go test` |
| [`processor_test.go`](../../internal/agent/processor_test.go) | Fake-provider loop tests: ordered parallel results, mode gates, cancellation, reporting, redaction, fan-out, tasks. | `Test*` | `go test` |
| [`adddir_test.go`](../../internal/agent/adddir_test.go) | Validates granted roots and their system-prompt representation. | `Test*` | `go test` |
| [`redaction_security_test.go`](../../internal/agent/redaction_security_test.go) | Checks queued/restored secrets and task-plan goal redaction. | `Test*` | `go test` |
| [`subagent_security_test.go`](../../internal/agent/subagent_security_test.go) | Checks enforced read-only exploration and complete, visible command approval. | `Test*` | `go test` |

## internal-config

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`config.go`](../../internal/config/config.go#L14) | Defines defaults, merges trusted global settings with permitted project preferences, enforces scoped writes, offers dotted-path accessors. | `Config`, `LLM`, `Features`, `Guardrails`; `Default`, `Load`, `Set`, `SetGlobalLLM`, `Path`, `GlobalPath`, `IgnoredProjectKeys`, `GetByPath`, `SetByPath`, `ParseValue` and config accessors | `cmd`, `agent`, `llm`, `tools`, `tui` |
| [`config_test.go`](../../internal/config/config_test.go) | Covers defaults/accessors, parsing, trust-aware merging, ignored-key notices, and context-window preferences. | `Test*` | `go test` |

## internal-creds

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`creds.go`](../../internal/creds/creds.go#L54) | Reads/writes provider API keys in the private user credentials JSON file. | `APIKey`, `SetAPIKey`, `Path` | Provider credential resolution; TUI `/connect` |
| [`creds_test.go`](../../internal/creds/creds_test.go) | Exercises stored-key round-trip. | `Test*` | `go test` |

## internal-llm

Wire formats are internal to each adapter; the agent consumes one common message/request contract.

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`llm.go`](../../internal/llm/llm.go#L19) | Defines messages, tool calls, requests, provider/stream interfaces; resolves backend, credentials and endpoint policy. | `Role`, `ToolCall`, `Usage`, `Message`, `ToolDef`, `Request`, `Provider`, `Streamer`, `MissingKeyError`; `New`, `InferProvider`, `ResolveProvider` | `agent`, `cmd`, `tui`; adapter implementations |
| [`openai.go`](../../internal/llm/openai.go#L8) | Converts common requests/results to OpenAI Chat Completions, including streamed tool-call assembly. | Private `openaiProvider` implementing `Name`, `Complete`, `Stream` | `llm.New`; agent through interfaces |
| [`anthropic.go`](../../internal/llm/anthropic.go#L8) | Converts common messages to Anthropic content blocks and merges adjacent tool-result turns; handles completion/stream responses. | Private `anthropicProvider` implementing `Name`, `Complete`, `Stream` | `llm.New`; agent through interfaces |
| [`google.go`](../../internal/llm/google.go#L9) | Converts requests to Gemini content/function declarations and translates complete/streamed responses. | Private `googleProvider` implementing `Name`, `Complete`, `Stream` | `llm.New`; agent through interfaces |
| [`http.go`](../../internal/llm/http.go#L22) | Endpoint construction, bounded JSON HTTP responses, retries/backoff, SSE transport and parser, redacted errors. | Package-private `providerEndpoint`, `postJSON`, `streamSSE`, `scanSSE` | All provider adapters |
| [`context_window.go`](../../internal/llm/context_window.go#L14) | Longest-prefix lookup of known model context-window sizes; unknown models return no value. | `ContextWindow` | Processor status calculations |
| [`pricing.go`](../../internal/llm/pricing.go#L6) | Longest-prefix model price lookup and token-cost estimate. | `Price`, `PriceFor`, `Cost` | Processor status calculations |
| [`llm_test.go`](../../internal/llm/llm_test.go) | Provider inference, Gemini schema conversion, endpoint/key policy, reasoning effort. | `Test*` | `go test` |
| [`http_test.go`](../../internal/llm/http_test.go) | Transient retry, client-error behavior, cancellation and redacted errors. | `Test*` | `go test` |
| [`stream_test.go`](../../internal/llm/stream_test.go) | SSE parsing and each provider's text/tool-call stream assembly, including custom OpenAI-compatible endpoint. | `Test*` | `go test` |
| [`context_window_test.go`](../../internal/llm/context_window_test.go), [`pricing_test.go`](../../internal/llm/pricing_test.go) | Longest-prefix metadata matching and unknown-model behavior. | `Test*` | `go test` |

## internal-memory

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`memory.go`](../../internal/memory/memory.go#L10) | Loads nonempty project `COOLCODE.md`, falling back to user-global instructions. | `File`, `LoadProjectInstructions` | Agent context construction |

## internal-project

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`gitignore.go`](../../internal/project/gitignore.go#L14) | Converts a subset of root `.gitignore` rules to a checker; ignores negation rules. | `GitIgnoreChecker`, `NewGitIgnoreChecker` | Agent startup/context and tool context |
| [`tree.go`](../../internal/project/tree.go#L19) | Produces a sorted ASCII directory tree with supplied ignore predicate and depth limit. | `FolderStructure` | Agent context construction/refresh |
| [`scan.go`](../../internal/project/scan.go#L12) | Detects known entrypoint files, package.json frameworks/scripts and language extensions. | `Scan`, `ScanProject` | CLI scan command; `project_summary` tool |
| [`gitignore_test.go`](../../internal/project/gitignore_test.go) | Checks common ignore patterns and missing-file behavior. | `Test*` | `go test` |

## internal-security

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`redact.go`](../../internal/security/redact.go#L11) | Replaces common secret formats and matching sensitive environment-variable values. | `Replacement`, `Redact` | Agent context/processor/planning/subagents, tool execution/approval, provider errors |
| [`sanitize.go`](../../internal/security/sanitize.go#L14) | Strips terminal escapes and control characters; provides single-line presentation. | `SanitizeTerminal`, `SanitizeLine` | TUI rendering and command output |
| [`redact_test.go`](../../internal/security/redact_test.go), [`sanitize_test.go`](../../internal/security/sanitize_test.go) | Covers secret-format masking, escape removal, and visible one-line payloads. | `Test*` | `go test` |

## internal-session

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`session.go`](../../internal/session/session.go#L16) | JSON snapshots under the user's sessions directory; UUID IDs, private modes, directory filtering and newest-first ordering. | `Data`, `NewID`, `Save`, `Load`, `List`, `Latest` | CLI resume startup; TUI persistence/session picker |
| [`session_test.go`](../../internal/session/session_test.go) | Checks file privacy and traversal-resistant session IDs. | `Test*` | `go test` |

## internal-skills

This is the application's Go skill installer and loader. Its destinations are `.coolcode/skills` and `~/.coolcode/skills`, as defined in the files below.

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`skills.go`](../../internal/skills/skills.go#L14) | Parses `SKILL.md` name/description/body, discovers project/global directories with project precedence, supplies catalog and named body. | `Skill`, `ParseFile`, `Discover`, `Catalog`, `Body`, `Names` | Agent context; `use_skill` tool; installer naming |
| [`installer.go`](../../internal/skills/installer.go#L92) | Accepts local file/directory or shallow Git clone; locates skills, skips symlinks, stages copies and replaces destinations using `os.Root` with backup restoration. | `InstallResult`, `IsGitURL`, `Install` | CLI `skill install`; TUI `/install-skill` |
| [`skills_test.go`](../../internal/skills/skills_test.go) | Parsing, discovery, installation, source-symlink exclusions. | `Test*` | `go test` |
| [`installer_security_test.go`](../../internal/skills/installer_security_test.go) | Destination symlink rejection and preservation of previous skill after copy failure. | `Test*` | `go test` |

## internal-tools

[`registry.go`](../../internal/tools/registry.go#L15) publishes 24 named tools. Their definitions are package-private values registered in `All`; the agent separately adds `spawn_agent` and `update_task_list`. Tool implementations share `Context` and return `types.ToolResult`.

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`tools.go`](../../internal/tools/tools.go#L19) | Capability metadata, ambient roots/config/cancellation context, schema helpers, untrusted-content wrapper. | `Context`, `Tool`, `Untrusted`; `Context.Roots`, `Context.Context` | Agent dispatch; all tool implementations |
| [`registry.go`](../../internal/tools/registry.go#L15) | Registers tools, looks up/executes calls, enforces read-only dispatch, prepares exact command approval and edit previews. | `All`, `Lookup`, `Run`, `RunReadOnly`, `DangerReason`, `EditPreview`, `IsMutating`, `IsReadOnly`, `ReadOnlyTools`, `PrepareCommand` | Processor and subagent loops |
| [`fs_tools.go`](../../internal/tools/fs_tools.go#L16) | Read/range-read, literal edit, create, rename, recent-file listing, bulk replace, TypeScript module scaffolding. | No Go exports; tools `read_file`, `open_file_at`, `edit_file`, `new_file`, `rename_file`, `list_recent_files`, `replace_in_files`, `new_module` | Registry dispatch |
| [`search_tools.go`](../../internal/tools/search_tools.go#L18) | Glob search, in-process regex content search, policy-filtered external ripgrep symbol search. | No Go exports; tools `glob`, `grep`, `find_symbol` | Registry dispatch |
| [`shell_tools.go`](../../internal/tools/shell_tools.go#L11) | Shell commands, detected/explicit tests and lint fixes, per-file formatting, package.json script changes. | No Go exports; tools `shell_command`, `run_tests`, `lint_fix`, `format_file`, `add_script` | Registry dispatch |
| [`git_tools.go`](../../internal/tools/git_tools.go#L12) | Git status, policy-filtered diffs, staging and commits using direct argument vectors. | No Go exports; tools `git_status`, `git_diff`, `git_commit` | Registry dispatch |
| [`misc_tools.go`](../../internal/tools/misc_tools.go#L14) | Lightweight project summary, README section append/create, named skill-body loading. | No Go exports; tools `project_summary`, `generate_readme_section`, `use_skill` | Registry dispatch |
| [`web_tools.go`](../../internal/tools/web_tools.go#L32) | HTTPS page retrieval/text extraction and DuckDuckGo HTML search with bounded results. | No Go exports; tools `web_fetch`, `web_search` | Registry dispatch |
| [`safety.go`](../../internal/tools/safety.go#L16) | Canonical path/root checks, case-insensitive read policy and protected-write rules; path/manifest helpers. | `BlockedPath`, `EnsureAbsoluteWithinRoot`, `EnsureAbsoluteWithinRoots`, `ResolveWithinRoots`, `ValidateReadPath`, `ResolveReadPath`, `ResolveWritePath` | Tool implementations; agent pins/context/root grants |
| [`network_safety.go`](../../internal/tools/network_safety.go#L12) | HTTPS URL checks, restricted IP resolution/dialing, redirect checks, no-proxy web client. | Package-private `safeWebClient`, `validateWebURL`, `safeDialContext`, `blockedNetworkIP` | Web tools |
| [`shell.go`](../../internal/tools/shell.go#L25) | Context-cancelled subprocess execution, timeouts, stdout/stderr capture/redaction and environment allowlist. | Package-private `execCommand`, `execArgv`, `runCommandRaw`, `commandTimeout` | Shell/Git/search tool implementations |
| [`glob_util.go`](../../internal/tools/glob_util.go#L21) | Shared glob/walk filtering; file lists for primary and additional-root completion. | `ProjectFiles`, `ProjectFilesAbs`; private `globFiles`, `walkFiles` | File/search tools; TUI completion |
| [`lang_detect.go`](../../internal/tools/lang_detect.go#L29) | Infers Go/Rust/Node/Python test and lint commands and extension-specific formatter argv. | Package-private `detectProject`, `defaultTestCommand`, `defaultLintFixCommand`, `formatFileArgv` | Command preparation and shell tools |
| [`fs_tools_test.go`](../../internal/tools/fs_tools_test.go) | Basic file/search operations, root and symlink rejection, guardrails, dry-run default. | `Test*` | `go test` |
| [`git_tools_test.go`](../../internal/tools/git_tools_test.go) | Filtered diffs, literal argv handling, status, application-pattern semantics. | `Test*` | `go test` |
| [`lang_detect_test.go`](../../internal/tools/lang_detect_test.go) | Project-marker and default-command detection. | `Test*` | `go test` |
| [`safety_test.go`](../../internal/tools/safety_test.go) | Paths, multiple roots, previews, permissions, symlink/parent semantics and cycles. | `Test*` | `go test` |
| [`guards_test.go`](../../internal/tools/guards_test.go) | Protected paths, search filtering, mode preservation, shell working directories, range reads. | `Test*` | `go test` |
| [`security_test.go`](../../internal/tools/security_test.go) | Environment filtering, network policy, safe argv, confirmations, timeouts. | `Test*` | `go test` |
| [`approval_security_test.go`](../../internal/tools/approval_security_test.go) | Full visible commands, frozen inferred commands, rejection of redaction-hidden syntax. | `Test*` | `go test` |
| [`case_security_test.go`](../../internal/tools/case_security_test.go) | Case aliases in path guards and Git diff filtering. | `Test*` | `go test` |
| [`module_security_test.go`](../../internal/tools/module_security_test.go) | Validates all module destinations before writes and checks scaffold contents. | `Test*` | `go test` |
| [`readme_security_test.go`](../../internal/tools/readme_security_test.go) | README write policy plus create/append behavior. | `Test*` | `go test` |
| [`rename_security_test.go`](../../internal/tools/rename_security_test.go) | Rejects directory moves while retaining regular-file rename behavior. | `Test*` | `go test` |
| [`search_security_test.go`](../../internal/tools/search_security_test.go) | Checks each ripgrep candidate through directory aliases. | `Test*` | `go test` |

## internal-tui

The only exported constructor boundary is `Run(RunOptions)`. Bubble Tea's model methods and the reporter bridge operate on package-private types.

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`run.go`](../../internal/tui/run.go#L10) | Creates model, restores visible transcript, wires reporter/confirmation bridge, starts alternate-screen program. | `RunOptions`, `Run` | `cmd.runInteractive` |
| [`app.go`](../../internal/tui/app.go#L26) | UI state and message types; processor bridge; transcript/stream rendering caches; snapshot persistence. | Private `model`, `bridge`, `newModel`, `persist`; `model.Init` and reporter methods | `Run`; update/render/command files |
| [`update.go`](../../internal/tui/update.go#L20) | Bubble Tea event reducer: keys, menus, approvals, streamed messages, query launch, cancellation, completion, credential connection and session resume. | `model.Update` on private type; private event handlers | Bubble Tea runtime |
| [`dispatch.go`](../../internal/tui/dispatch.go#L20) | Executes slash commands for mode, provider preferences, roots, pins, context, compaction, skills, sessions and quit. | Package-private `dispatchCommand`, `switchModel`, `helpText` | Input submission |
| [`commands.go`](../../internal/tui/commands.go#L28) | Slash-command catalog, mode cycle, command/file suggestions and cached file enumeration. | Package-private `commands`, `nextMode`, `matchCommands`, `atToken`, `matchFiles` | Event handlers and command help |
| [`layout.go`](../../internal/tui/layout.go#L16) | Computes responsive transcript/sidebar/footer geometry from terminal dimensions. | Package-private `layout`, `computeLayout` | Resize handling and view composition |
| [`view.go`](../../internal/tui/view.go#L22) | Composes transcript, header/sidebar, tasks, status, composer, suggestions and confirmation/menu overlays. | `model.View` on private type; private rendering helpers | Bubble Tea runtime; transcript helpers |
| [`markdown.go`](../../internal/tui/markdown.go#L24) | Cached width-specific Glamour Markdown renderers and palette-driven syntax styling. | Package-private `renderMarkdown`, `markdownStyle` | Transcript rendering |
| [`styles.go`](../../internal/tui/styles.go#L10) | Adaptive light/dark palette and reusable Lip Gloss styles, task/mode glyph styling. | Package-private style values, `taskGlyph`, `modeStyle` | TUI rendering and input initialization |
| [`app_test.go`](../../internal/tui/app_test.go) | Input, mode changes, help/pins/effort, exit and suggestions. | `Test*` | `go test` |
| [`commands_test.go`](../../internal/tui/commands_test.go) | `@` token extraction and file-match ranking. | `Test*` | `go test` |
| [`adddir_test.go`](../../internal/tui/adddir_test.go) | Additional-directory slash command and completion integration. | `Test*` | `go test` |
| [`view_test.go`](../../internal/tui/view_test.go) | Responsive sizes, bounded overlays, sanitized output, styling, transcript/stream caches, task/status presentation. | `Test*` | `go test` |

## internal-types

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`types.go`](../../internal/types/types.go#L5) | Shared operation modes, tool return contract and task-list state. | `AgentMode`, `ModePlan`, `ModeAgent`, `ModeAsk`, `ToolResult`, `TaskStatus`, `TaskItem`, `TaskList` | `agent`, `tools`, `cmd`, `tui` |

## github

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [`ci.yml`](../../.github/workflows/ci.yml#L1) | Formatting, vet, race tests, static security and vulnerability checks, build/version smoke test. | `build` job | Push/PR targeting `master` or `main` |
| [`release.yml`](../../.github/workflows/release.yml#L1) | Cross-builds five OS/architecture targets; creates archives/checksums/SBOMs/signatures and publishes completed draft with README-derived notes. | `release` matrix and `publish` jobs | `v*` tag push |
| [`dependabot.yml`](../../.github/dependabot.yml#L1) | Weekly Go-module and GitHub Actions update checks. | Two update definitions | Dependabot |

## Excluded

[`go.sum`](../../go.sum) is a dependency checksum file, not application logic. [`.gitignore`](../../.gitignore) is repository housekeeping, and [`.env.example`](../../.env.example) is a configuration example; their contents are not reproduced here. Generated executables, caches, local credentials/settings/sessions, vendored dependencies, and this guide's own files are not source-map rows. No generated or vendored implementation appeared in the tracked source inventory used for this guide.
