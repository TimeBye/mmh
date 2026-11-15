package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallTo(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "mmh_darwin_arm64")
	if err := os.WriteFile(src, []byte("fake-binary"), 0755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	dir := filepath.Join(root, "bin")
	if err := installTo(src, dir); err != nil {
		t.Fatalf("installTo: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "mmh"))
	if err != nil {
		t.Fatalf("read installed binary: %v", err)
	}
	if string(data) != "fake-binary" {
		t.Fatalf("installed binary content = %q, want fake-binary", data)
	}

	want, err := filepath.EvalSymlinks(filepath.Join(dir, "mmh"))
	if err != nil {
		t.Fatalf("resolve installed binary: %v", err)
	}
	for _, name := range symlinkCommands {
		link := filepath.Join(dir, name)
		info, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("lstat %s: %v", link, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("%s is not a symlink", link)
		}
		resolved, err := filepath.EvalSymlinks(link)
		if err != nil {
			t.Fatalf("resolve %s: %v", link, err)
		}
		if resolved != want {
			t.Fatalf("%s resolves to %s, want %s", link, resolved, want)
		}
	}
}

func TestInstallToDefaultDir(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "mmh_darwin_arm64")
	if err := os.WriteFile(src, []byte("fake-binary"), 0755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	if err := installTo(src, ""); err != nil {
		t.Fatalf("installTo: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "mmh")); err != nil {
		t.Fatalf("install to binary dir: %v", err)
	}
	want, err := filepath.EvalSymlinks(filepath.Join(root, "mmh"))
	if err != nil {
		t.Fatalf("resolve installed binary: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(filepath.Join(root, "mcs")); err != nil || resolved != want {
		t.Fatalf("mcs resolves to %s (%v), want %s", resolved, err, want)
	}
}

func TestPathContainsSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(t.TempDir(), "bin")
	if err := os.Symlink(dir, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	t.Setenv("PATH", "/usr/bin:"+link+":/bin")
	if !pathContains(dir) {
		t.Fatalf("pathContains(%s) with PATH containing symlink %s = false, want true", dir, link)
	}
}

func TestPathContainsCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	alt := strings.ToUpper(dir)
	if _, err := os.Stat(alt); err != nil {
		t.Skip("case-sensitive filesystem")
	}

	t.Setenv("PATH", "/usr/bin:"+alt+":/bin")
	if !pathContains(dir) {
		t.Fatalf("pathContains(%s) with differently cased PATH entry %s = false, want true", dir, alt)
	}
}

func TestPathContains(t *testing.T) {
	dir := t.TempDir()
	other := t.TempDir()

	t.Setenv("PATH", "/usr/bin:"+dir+":/bin")
	if !pathContains(dir) {
		t.Fatalf("pathContains(%s) = false, want true", dir)
	}
	if pathContains(other) {
		t.Fatalf("pathContains(%s) = true, want false", other)
	}
}
