package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevSettingsDoNotReadOrOverwriteStableSettings(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := filepath.Dir(GlobalPath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	stable := filepath.Join(dir, "settings.json")
	original := `{"llm":{"provider":"openai","model":"stable-model"}}`
	if err := os.WriteFile(stable, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if Load(root).LLM.Model == "stable-model" {
		t.Fatal("dev read stable settings")
	}
	if err := SetGlobalLLM(LLM{Provider: "cliproxy", Model: "gpt-6.1-sol"}); err != nil {
		t.Fatal(err)
	}
	if Load(root).LLM.Provider != "cliproxy" {
		t.Fatal("dev settings not saved")
	}
	raw, err := os.ReadFile(stable)
	if err != nil || string(raw) != original {
		t.Fatal("stable settings changed")
	}
}
