package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvImportsKeysButNotEndpointsOrProcessControls(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, name := range []string{"CLIPROXY_API_KEY", "COOLCODE_API_BASE_URL", "HTTP_PROXY"} {
		t.Setenv(name, "")
		_ = os.Unsetenv(name)
	}
	data := []byte("CLIPROXY_API_KEY=proxy-key\nCOOLCODE_API_BASE_URL=https://evil.example/v1\nHTTP_PROXY=https://evil.example\n")
	if err := os.WriteFile(filepath.Join(".", ".env"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	loadEnv()
	if os.Getenv("CLIPROXY_API_KEY") != "proxy-key" {
		t.Fatal("API key was not loaded")
	}
	if os.Getenv("COOLCODE_API_BASE_URL") != "" || os.Getenv("HTTP_PROXY") != "" {
		t.Fatal("untrusted .env changed endpoint or proxy settings")
	}
}

// TestLoadEnvIgnoresCoolcodeAPIKey covers credential substitution from a
// repository .env. COOLCODE_API_KEY is consulted ahead of the stored /connect
// credential and needs no global setting, so honouring it from a project file
// let a cloned repository decide which provider account every request was
// billed and logged against.
func TestLoadEnvIgnoresCoolcodeAPIKey(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, name := range []string{"COOLCODE_API_KEY", "CLIPROXY_API_KEY"} {
		t.Setenv(name, "")
		_ = os.Unsetenv(name)
	}
	body := "COOLCODE_API_KEY=sk-attacker\nCLIPROXY_API_KEY=sk-proxy\nPATH=/evil\n"
	if err := os.WriteFile(filepath.Join(".", ".env"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	loadEnv()

	if got := os.Getenv("COOLCODE_API_KEY"); got != "" {
		t.Errorf("COOLCODE_API_KEY was imported from .env: %q", got)
	}
	// A named key stays available: it is inert unless the user's own global
	// llm.apiKeyEnv points at it.
	if got := os.Getenv("CLIPROXY_API_KEY"); got != "sk-proxy" {
		t.Errorf("CLIPROXY_API_KEY = %q, want sk-proxy", got)
	}
	if got := os.Getenv("PATH"); got == "/evil" {
		t.Error("PATH was imported from .env")
	}
}

// TestSameDirResolvesSymlinks guards the --resume check, which refuses a
// session recorded elsewhere because its /add-dir grants are replayed on
// restore. It must not reject a legitimate resume just because the path is
// spelled through a symlink.
func TestSameDirResolvesSymlinks(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if !sameDir(real, link) {
		t.Error("symlinked spelling of the same directory was rejected")
	}
	if !sameDir(real, real) {
		t.Error("identical paths compared unequal")
	}
	if sameDir(real, t.TempDir()) {
		t.Error("different directories compared equal")
	}
	if sameDir("", real) {
		t.Error("empty path compared equal")
	}
}
