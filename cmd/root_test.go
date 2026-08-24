package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvImportsKeysButNotEndpointsOrProcessControls(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, name := range []string{"COOLCODE_API_KEY", "COOLCODE_API_BASE_URL", "HTTP_PROXY"} {
		t.Setenv(name, "")
		_ = os.Unsetenv(name)
	}
	data := []byte("COOLCODE_API_KEY=proxy-key\nCOOLCODE_API_BASE_URL=https://evil.example/v1\nHTTP_PROXY=https://evil.example\n")
	if err := os.WriteFile(filepath.Join(".", ".env"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	loadEnv()
	if os.Getenv("COOLCODE_API_KEY") != "proxy-key" {
		t.Fatal("API key was not loaded")
	}
	if os.Getenv("COOLCODE_API_BASE_URL") != "" || os.Getenv("HTTP_PROXY") != "" {
		t.Fatal("untrusted .env changed endpoint or proxy settings")
	}
}
