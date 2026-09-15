package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanCacheNeverClobbersLinkedTargets(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink"} {
		for _, refresh := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: "/normal", true: "/refresh"}[refresh], func(t *testing.T) {
				root := t.TempDir()
				t.Chdir(root)
				t.Setenv("HOME", t.TempDir())
				sentinel := filepath.Join(t.TempDir(), "sentinel")
				if err := os.WriteFile(sentinel, []byte("unchanged"), 0600); err != nil {
					t.Fatal(err)
				}
				link := os.Link
				if kind == "symlink" {
					link = os.Symlink
				}
				if err := link(sentinel, filepath.Join(root, scanCacheName)); err != nil {
					t.Fatal(err)
				}
				c := scanCmd()
				if refresh {
					c.SetArgs([]string{"--refresh"})
				} else {
					c.SetArgs(nil)
				}
				if err := c.Execute(); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(sentinel)
				if err != nil || string(data) != "unchanged" {
					t.Fatalf("sentinel changed: %q %v", data, err)
				}
			})
		}
	}
}

func TestScanCacheNormalRoundTrip(t *testing.T) {
	root := t.TempDir()
	for _, data := range []string{`{"root":"first"}`, `{"root":"second"}`} {
		if err := writeScanCache(root, []byte(data)); err != nil {
			t.Fatal(err)
		}
		got, err := readScanCache(root)
		if err != nil || string(got) != data {
			t.Fatalf("%q %v", got, err)
		}
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 1 {
		t.Fatal("temporary cache entries left behind")
	}
}
