package repstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsCorpus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rep.json")
	if err := os.WriteFile(path, []byte(`{"abc":{"hash":"abc","trusters":12,"firstSeen":"2026-04-01T00:00:00Z"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	src, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	sig, ok := src.Lookup("abc")
	if !ok || sig.Trusters != 12 {
		t.Errorf("corpus not loaded: %+v ok=%v", sig, ok)
	}
}

func TestLoadMissingFileIsSilentNoOp(t *testing.T) {
	src, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("missing file must not error: %v", err)
	}
	if _, ok := src.Lookup("anything"); ok {
		t.Errorf("missing corpus should resolve nothing")
	}
}

func TestLoadEmptyPath(t *testing.T) {
	src, err := Load("")
	if err != nil || src == nil {
		t.Fatalf("empty path should be an empty, non-nil corpus: %v", err)
	}
	if len(src) != 0 {
		t.Errorf("empty path should load nothing, got %d", len(src))
	}
}
