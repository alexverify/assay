package resolve

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// makeTarGz builds an in-memory .tar.gz from path->content entries, rooted at
// "package/" like a real npm tarball.
func makeTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: "package/" + name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "pkg.tgz")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// makeTarGzAt writes a package-rooted .tar.gz to dir/filename and returns its path.
func makeTarGzAt(t *testing.T, dir, filename string, files map[string]string) string {
	t.Helper()
	src := makeTarGz(t, files)
	dst := filepath.Join(dir, filename)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestExtractTarGzStripsPackagePrefix(t *testing.T) {
	src := makeTarGz(t, map[string]string{
		"package.json":  `{"name":"x"}`,
		"dist/index.js": "console.log(1)",
	})
	dest := t.TempDir()
	if err := extractTarGz(src, dest); err != nil {
		t.Fatalf("extractTarGz: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "dist", "index.js"))
	if err != nil || string(got) != "console.log(1)" {
		t.Fatalf("extracted file wrong: %q err=%v", got, err)
	}
}

func TestExtractTarGzRejectsPathTraversal(t *testing.T) {
	// Craft a tarball with a ../ escape.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: "package/../evil.sh", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg}
	_ = tw.WriteHeader(hdr)
	_, _ = tw.Write([]byte("x"))
	_ = tw.Close()
	_ = gz.Close()
	src := filepath.Join(t.TempDir(), "evil.tgz")
	if err := os.WriteFile(src, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := extractTarGz(src, t.TempDir()); err == nil {
		t.Fatal("expected path-traversal rejection")
	}
}
