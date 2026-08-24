# Security policy

## Reporting a vulnerability

Please report vulnerabilities privately through GitHub Security Advisories:

https://github.com/rushikeshg25/cool-code/security/advisories/new

Do not include credentials, private source code, or exploit data in a public issue.

## Trust model

Cool-Code sends prompts and explicitly accessed project context to the configured model provider. Provider identity, endpoints, credential selection, guardrails, and confirmation bypasses are trusted user settings and cannot be selected by repository configuration.

Plan and Ask modes are read-only. Agent mode can modify explicitly granted workspace roots. Arbitrary shell commands, project-code execution, and git commits require confirmation unless the user explicitly enables the global danger bypass.

Web fetching blocks private and metadata networks. Child commands do not inherit API keys, tokens, cookies, cloud credentials, or proxy variables. Common secret formats are redacted before provider egress and session persistence.

These controls reduce accidental disclosure but do not make model providers or explicitly approved commands trusted. Review provider retention policies and confirmation prompts before using the CLI with sensitive repositories.

## Release integrity

Release archives include SHA-256 checksums, SPDX JSON SBOMs, and keyless Sigstore bundles. GitHub Actions dependencies are pinned to full commit hashes.
