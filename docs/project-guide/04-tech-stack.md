# Tech Stack

## Languages and runtimes

| Language or runtime | Version | Pinned at |
| --- | --- | --- |
| Go | Module declares `1.25.13`; CI and releases select the toolchain from this declaration. | [`go.mod:3`](../../go.mod#L3), [CI:22](../../.github/workflows/ci.yml#L22), [release:34](../../.github/workflows/release.yml#L34) |
| Bash | Host executable; no version pin. Used for approved command strings, including inferred test/lint commands. | [`shell.go:23`](../../internal/tools/shell.go#L23) |

The application is one Go module with a local CLI entry point, rather than a Node application. Node and other language toolchains can still be runtime dependencies of tools operating on the user's project; see the external commands below. [Module](../../go.mod#L1) · [entry](../../main.go#L1) · [toolchain detection](../../internal/tools/lang_detect.go#L25).

## Frameworks and major libraries

All direct dependencies are listed here. Provider HTTP and SSE handling use the Go standard library rather than a provider SDK. [HTTP implementation](../../internal/llm/http.go#L3).

| Library | Version and manifest pin | Used for | Used in |
| --- | --- | --- | --- |
| Cobra | [`v1.10.2`](../../go.mod#L15) | CLI command tree and flags | [`cmd/root.go`](../../cmd/root.go#L12) |
| Bubble Tea | [`v1.3.10`](../../go.mod#L9) | Terminal application event loop | [`internal/tui/run.go`](../../internal/tui/run.go#L4) |
| Bubbles | [`v1.0.0`](../../go.mod#L8) | Spinner, textarea, text input, viewport | [`internal/tui/app.go`](../../internal/tui/app.go#L10) |
| Lip Gloss | [`v1.1.1-0.20250404203927-76690c660834`](../../go.mod#L11) | Terminal styling and layout | [`internal/tui/styles.go`](../../internal/tui/styles.go#L4) |
| Glamour | [`v1.0.0`](../../go.mod#L10) | Markdown rendering with width-specific cached renderers | [`internal/tui/markdown.go`](../../internal/tui/markdown.go#L14) |
| Charm `x/ansi` | [`v0.11.6`](../../go.mod#L12) | Terminal text width and ANSI-aware handling | [`internal/tui/view.go`](../../internal/tui/view.go#L11), [`markdown.go`](../../internal/tui/markdown.go#L11) |
| Doublestar v4 | [`v4.10.0`](../../go.mod#L7) | Recursive glob matching for search and read policy | [`internal/tools/glob_util.go`](../../internal/tools/glob_util.go#L9), [`safety.go`](../../internal/tools/safety.go#L10) |
| Atotto clipboard | [`v0.1.4`](../../go.mod#L6) | Optional copying of completed answers | [`internal/tui/update.go`](../../internal/tui/update.go#L135) |
| Google UUID | [`v1.6.0`](../../go.mod#L13) | Session IDs, skill staging names and scan-cache temporary names | [`internal/session/session.go`](../../internal/session/session.go#L33), [`internal/skills/installer.go`](../../internal/skills/installer.go#L14), [`cmd/scan.go`](../../cmd/scan.go#L9) |
| Godotenv | [`v1.5.1`](../../go.mod#L14) | Parse a regular local `.env`; the caller selectively imports allowed key names | [`cmd/root.go`](../../cmd/root.go#L169) |

The Lip Gloss version is a pseudo-version pinned to a particular commit. The manifest does not explain why that commit was selected. Indirect rendering and terminal dependencies are recorded separately in [`go.mod:18`](../../go.mod#L18); they are not application-level service boundaries.

## Data and infrastructure

| Service or storage | Role | Configured or implemented at |
| --- | --- | --- |
| Anthropic Messages API | Model completions and streaming | [Default `/v1/messages` endpoint](../../internal/llm/http.go#L30), [adapter](../../internal/llm/anthropic.go#L1) |
| OpenAI Chat Completions API | Model completions and streaming, including compatible proxies | [Default `/v1/chat/completions` endpoint](../../internal/llm/http.go#L31), [adapter](../../internal/llm/openai.go#L1) |
| Google Generative Language API | Gemini completions and streaming | [Default `/v1beta/models/` endpoint](../../internal/llm/http.go#L32), [adapter](../../internal/llm/google.go#L1) |
| Optional custom provider endpoint | Trusted config or process environment selects the endpoint; explicit proxy credential required | [`llm.go:167`](../../internal/llm/llm.go#L167), [`llm.go:211`](../../internal/llm/llm.go#L211) |
| DuckDuckGo HTML search | Web search result titles, URLs and snippets parsed from HTML; not a versioned search SDK | [`web_tools.go:109`](../../internal/tools/web_tools.go#L109), [endpoint](../../internal/tools/web_tools.go#L151) |
| Local JSON files | Settings at `~/.coolcode/settings.json`, credentials at `~/.coolcode/credentials.json`, session snapshots at `~/.coolcode/sessions/<uuid>.json` | [`config.go:80`](../../internal/config/config.go#L80), [`creds.go:20`](../../internal/creds/creds.go#L20), [`session.go:28`](../../internal/session/session.go#L28) |
| Project JSON files | Safe project preferences in `.coolcode.json`; optional scan cache in `.coolcode.scan.json` | [`config.go:75`](../../internal/config/config.go#L75), [`scan.go:80`](../../cmd/scan.go#L80) |
| Local Markdown | `COOLCODE.md` project memory and project/global skill directories | [`memory.go:13`](../../internal/memory/memory.go#L13), [`installer.go:90`](../../internal/skills/installer.go#L90) |

These are local-file stores and outbound services. Session persistence uses JSON marshaling and filesystem calls, not a database driver. Releases distribute binaries; their workflow does not provision an application server. [Persistence](../../internal/session/session.go#L45) · [release packaging](../../.github/workflows/release.yml#L36).

## Tooling

| Tool | Role | Configured at |
| --- | --- | --- |
| Make / Go build | `make build` and `make install` add `-trimpath`, strip debug symbols, and inject `cmd.Version` from Git | [`Makefile:1`](../../Makefile#L1) |
| Go test, vet and gofmt | Local `make test`, `make vet`, `make fmt`; CI checks formatting and runs tests with the race detector | [`Makefile:15`](../../Makefile#L15), [CI:25](../../.github/workflows/ci.yml#L25) |
| Gosec | Static security analysis; CI installs `v2.22.9`, then runs the documented exclusion list | [CI:36](../../.github/workflows/ci.yml#L36), [`Makefile:21`](../../Makefile#L21) |
| Govulncheck | Dependency/reachable vulnerability scan, CI installs `v1.7.0` | [CI:39](../../.github/workflows/ci.yml#L39) |
| GitHub Actions CI | Ubuntu job on pushes/PRs to `main` or `master`; read-only content permission; build and version smoke check | [CI:3](../../.github/workflows/ci.yml#L3) |
| GitHub Actions release | `v*` tag builds Linux amd64/arm64, macOS amd64/arm64 and Windows amd64 archives | [release:3](../../.github/workflows/release.yml#L3) |
| Anchore SBOM action | SPDX JSON per binary; action pinned to SHA, annotated `v0.24.0` | [release:52](../../.github/workflows/release.yml#L52) |
| Cosign installer / Cosign | Installer action pinned to SHA, annotated `v4.1.2`; signs each archive with a Sigstore bundle. The workflow does not explicitly pin the installed Cosign binary version. | [release:61](../../.github/workflows/release.yml#L61) |
| GitHub release action / `gh` | Upload draft artifacts, then publish after all matrix builds finish; `gh` version is supplied by the runner | [release:67](../../.github/workflows/release.yml#L67), [publish:77](../../.github/workflows/release.yml#L77) |
| Dependabot | Weekly Go module and GitHub Actions updates; at most five open PRs per ecosystem | [`.github/dependabot.yml:1`](../../.github/dependabot.yml#L1) |

Checkout and setup-go actions are pinned to commit SHAs, annotated as `v7.0.1` and `v7.0.0`. Local `make sec` assumes `gosec` already exists; only CI installs its version. [CI action pins](../../.github/workflows/ci.yml#L17) · [`Makefile:28`](../../Makefile#L28).

### External commands used by agent tools

These binaries are resolved on the host; their versions are not pinned by this repository.

| Command | When needed | Source |
| --- | --- | --- |
| `git` | Git status/diff/commit tools and remote skill installation | [`git_tools.go`](../../internal/tools/git_tools.go#L1), [`installer.go:120`](../../internal/skills/installer.go#L120) |
| `rg` | Symbol search | [`search_tools.go`](../../internal/tools/search_tools.go#L1) |
| `bash` | Approved shell strings and test/lint execution | [`shell.go:23`](../../internal/tools/shell.go#L23) |
| `go`, `gofmt` | Detected Go project tests, linting and formatting | [`lang_detect.go:60`](../../internal/tools/lang_detect.go#L60) |
| `cargo`, `rustfmt` | Detected Rust project tests, linting and formatting | [`lang_detect.go:60`](../../internal/tools/lang_detect.go#L60) |
| `pytest`, `ruff` | Detected Python project tests, linting and formatting | [`lang_detect.go:60`](../../internal/tools/lang_detect.go#L60) |
| `npm`, `npx prettier` | Node package scripts and fallback file formatter | [`lang_detect.go:42`](../../internal/tools/lang_detect.go#L42), [`lang_detect.go:98`](../../internal/tools/lang_detect.go#L98) |

## Notes

- The default model is `gemini-2.5-flash`, but model choice is configuration, not a dependency version pin. [Default config](../../internal/config/config.go#L55).
- The release job packages `.tar.gz`, checksums, an SBOM and a signature bundle for each target, including Windows. Cross-compilation is not evidence that all external-command workflows work on every target. [Release commands](../../.github/workflows/release.yml#L40) · [Bash dependency](../../internal/tools/shell.go#L30).
- The publishing job requires a nonempty README changelog section headed `### <version>:` matching the tag without `v`; a new release needs that section. [Release notes extraction](../../.github/workflows/release.yml#L88).

Next: [Decisions](05-decisions.md) · [Source map](03-structure.md).
