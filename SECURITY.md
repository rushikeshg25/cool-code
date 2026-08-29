# Security policy

## Reporting a vulnerability

Please report vulnerabilities privately through GitHub Security Advisories:

https://github.com/rushikeshg25/cool-code/security/advisories/new

Do not include credentials, private source code, or exploit data in a public issue.

## Trust model

Cool-Code sends prompts and explicitly accessed project context to the configured model provider. Provider identity, endpoints, credential selection, and confirmation bypasses are trusted user settings and cannot be selected by repository configuration. A repository may *add* read guardrail patterns, since that only narrows access; it can never remove one.

Repository-authored content (`COOLCODE.md`, `SKILL.md` files) and fetched web pages are wrapped in untrusted-content markers before they enter the prompt. Treat that as a mitigation, not a boundary: prompt injection is not solved, and a repository you do not trust should be opened in Plan or Ask mode.

Plan and Ask modes are read-only. Agent mode can modify explicitly granted workspace roots. Arbitrary shell commands, project-code execution, and git commits require confirmation unless the user explicitly enables the global danger bypass.

Web fetching blocks private, carrier-grade NAT, reserved, and metadata networks. Child commands do not inherit API keys, tokens, cookies, cloud credentials, or proxy variables, and commands built from model- or repository-controlled values are executed as argv with no shell. Common secret formats are redacted before provider egress and session persistence, and escape sequences are stripped from model output before it reaches the terminal.

Tool writes are refused inside `.git` and `.coolcode`, because a written hook or config turns a later git command into code execution.

These controls reduce accidental disclosure but do not make model providers or explicitly approved commands trusted. Review provider retention policies and confirmation prompts before using the CLI with sensitive repositories.

## Release integrity

Release archives include SHA-256 checksums, SPDX JSON SBOMs, and keyless Sigstore bundles. GitHub Actions dependencies are pinned to full commit hashes.
