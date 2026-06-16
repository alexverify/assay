package snapshotstore

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureWalksTextFilesSkipsBinaryAndGit(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "main.go"), []byte("package main\n"))
	mustWrite(t, filepath.Join(root, "sub", "a.py"), []byte("print(1)\n"))
	mustWrite(t, filepath.Join(root, "logo.png"), []byte{0x89, 0x50, 0x00, 0x01}) // binary → skipped
	mustWrite(t, filepath.Join(root, ".git", "config"), []byte("[core]\n"))       // .git → skipped

	s := New(t.TempDir())
	if err := s.Capture(context.Background(), "sha256-cap", root); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("sha256-cap")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["main.go"]; !ok {
		t.Errorf("main.go should be captured")
	}
	if _, ok := got["sub/a.py"]; !ok {
		t.Errorf("nested sub/a.py should be captured with a POSIX path")
	}
	if _, ok := got["logo.png"]; ok {
		t.Errorf("binary file must be skipped")
	}
	if _, ok := got[".git/config"]; ok {
		t.Errorf(".git contents must be skipped")
	}
}

func TestCaptureIdempotentByHash(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "f.txt"), []byte("v1\n"))
	s := New(t.TempDir())
	if err := s.Capture(context.Background(), "h", root); err != nil {
		t.Fatal(err)
	}
	// Change the file but reuse the hash: content-addressed, so Capture is a
	// no-op and the original bytes are preserved.
	mustWrite(t, filepath.Join(root, "f.txt"), []byte("v2-tampered\n"))
	if err := s.Capture(context.Background(), "h", root); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("h")
	if string(got["f.txt"]) != "v1\n" {
		t.Errorf("a stored hash must not be re-captured, got %q", got["f.txt"])
	}
}

func mustWrite(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPutGetRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	files := map[string][]byte{
		"main.go":       []byte("package main\nfunc main() {}\n"),
		"sub/helper.py": []byte("print('hi')\n"),
		"logo.bin":      {0x00, 0x01, 0x02, 0xff}, // binary survives via base64
	}
	if err := s.Put("sha256-abc123", files); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("sha256-abc123")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(files) {
		t.Fatalf("got %d files, want %d", len(got), len(files))
	}
	for path, want := range files {
		if !bytes.Equal(got[path], want) {
			t.Errorf("%s round-trip mismatch: got %q", path, got[path])
		}
	}
}

func TestGetMissingIsNilNoError(t *testing.T) {
	s := New(t.TempDir())
	got, err := s.Get("sha256-never-stored")
	if err != nil {
		t.Fatalf("a missing blob must not error: %v", err)
	}
	if got != nil {
		t.Errorf("a missing blob should be nil, got %v", got)
	}
	if s.Has("sha256-never-stored") {
		t.Errorf("Has should be false for an unstored hash")
	}
}

func TestGetMissingStoreDir(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "does-not-exist"))
	if got, err := s.Get("h"); err != nil || got != nil {
		t.Errorf("absent store should degrade to nil,nil; got %v, %v", got, err)
	}
}

func TestHasAfterPut(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Put("h1", map[string][]byte{"f": []byte("x")}); err != nil {
		t.Fatal(err)
	}
	if !s.Has("h1") {
		t.Errorf("Has should be true after Put")
	}
}

func TestPruneRemovesUnreferenced(t *testing.T) {
	s := New(t.TempDir())
	for _, h := range []string{"keep1", "keep2", "drop1", "drop2"} {
		if err := s.Put(h, map[string][]byte{"f": []byte(h)}); err != nil {
			t.Fatal(err)
		}
	}
	removed, err := s.Prune(map[string]bool{"keep1": true, "keep2": true})
	if err != nil {
		t.Fatal(err)
	}
	if removed != 2 {
		t.Errorf("Prune removed %d, want 2", removed)
	}
	if !s.Has("keep1") || !s.Has("keep2") {
		t.Errorf("kept blobs must survive prune")
	}
	if s.Has("drop1") || s.Has("drop2") {
		t.Errorf("unreferenced blobs must be removed")
	}
}

func TestEmptyHashIsNoOp(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Put("", map[string][]byte{"f": []byte("x")}); err != nil {
		t.Errorf("Put with empty hash should be a silent no-op, got %v", err)
	}
	if s.Has("") {
		t.Errorf("empty hash is not addressable")
	}
}
