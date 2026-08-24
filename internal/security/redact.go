// Package security contains data-loss-prevention helpers shared by the agent,
// tools, and provider transports.
package security

import (
	"os"
	"regexp"
	"strings"
)

const Replacement = "[REDACTED]"

var (
	namedSecretRE = regexp.MustCompile(`(?i)(api[_-]?key|access[_-]?token|auth[_-]?token|authorization|client[_-]?secret|password|passwd|private[_-]?key|secret)(["']?\s*[:=]\s*["']?)([^\s"',;}{]{4,})`)
	openAIKeyRE   = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{12,}\b`)
	githubKeyRE   = regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b`)
	awsKeyRE      = regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b`)
	pemBlockRE    = regexp.MustCompile(`(?s)-----BEGIN [^-\n]*PRIVATE KEY-----.*?-----END [^-\n]*PRIVATE KEY-----`)
)

// Redact removes common credential formats and the values of sensitive
// environment variables. It is intentionally applied before model egress and
// persistence, not merely before terminal rendering.
func Redact(input string) string {
	if input == "" {
		return input
	}
	out := input
	for _, item := range os.Environ() {
		name, value, ok := strings.Cut(item, "=")
		if !ok || len(value) < 8 || !sensitiveName(name) {
			continue
		}
		out = strings.ReplaceAll(out, value, Replacement)
	}
	out = namedSecretRE.ReplaceAllString(out, `${1}${2}`+Replacement)
	out = openAIKeyRE.ReplaceAllString(out, Replacement)
	out = githubKeyRE.ReplaceAllString(out, Replacement)
	out = awsKeyRE.ReplaceAllString(out, Replacement)
	out = pemBlockRE.ReplaceAllString(out, Replacement)
	return out
}

func sensitiveName(name string) bool {
	n := strings.ToUpper(name)
	for _, marker := range []string{
		"API_KEY", "TOKEN", "SECRET", "PASSWORD", "PASSWD", "PRIVATE_KEY",
		"CREDENTIAL", "AUTHORIZATION", "COOKIE",
	} {
		if strings.Contains(n, marker) {
			return true
		}
	}
	return false
}
