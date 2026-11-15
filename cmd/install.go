package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mritd/mmh/common"
	"github.com/spf13/cobra"
)

// symlinkCommands are the shortcut commands created on install
var symlinkCommands = []string{"mcp", "mcs", "mcx", "mec", "mgo", "mping", "mtun"}

var installDir string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install mmh to the target directory and create command symlinks",
	Run: func(cmd *cobra.Command, args []string) {
		exe, err := os.Executable()
		common.CheckAndExit(err)
		common.CheckAndExit(installTo(exe, installDir))
	},
}

func init() {
	installCmd.Flags().StringVar(&installDir, "dir", "", "install directory (default: directory of the running binary)")
	rootCmd.AddCommand(installCmd)
}

// installTo copies the current executable into dir as "mmh" and creates
// the shortcut command symlinks. When dir is empty, the directory containing
// the running binary is used
func installTo(exe, dir string) error {
	exe, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	if dir == "" {
		dir = filepath.Dir(exe)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	target := filepath.Join(dir, "mmh")
	if !sameFile(exe, target) {
		if err := copyExecutable(exe, target); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%w: run with sudo or use --dir to choose a writable directory", err)
			}
			return err
		}
	}

	for _, name := range symlinkCommands {
		link := filepath.Join(dir, name)
		if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove old symlink %s: %w", link, err)
		}
		if err := os.Symlink(target, link); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%w: run with sudo or use --dir to choose a writable directory", err)
			}
			return fmt.Errorf("create symlink %s: %w", link, err)
		}
	}

	fmt.Printf("installed => %s\n", target)
	for _, name := range symlinkCommands {
		fmt.Printf("symlink   => %s\n", filepath.Join(dir, name))
	}

	if !pathContains(dir) {
		fmt.Fprintf(os.Stderr, "warning: %s is not in PATH, run the following to use mmh directly:\n", dir)
		fmt.Fprintf(os.Stderr, "  export PATH=\"%s:$PATH\"\n", dir)
	}
	return nil
}

// pathContains reports whether dir is one of the PATH directories. Paths are
// compared by file identity when possible so that symlinks and case-insensitive
// filesystems (macOS) are handled correctly
func pathContains(dir string) bool {
	dir = normalizePath(dir)
	dirInfo, dirErr := os.Stat(dir)
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == "" {
			continue
		}
		p = normalizePath(p)
		if p == dir {
			return true
		}
		if dirErr == nil {
			if pInfo, err := os.Stat(p); err == nil && os.SameFile(dirInfo, pInfo) {
				return true
			}
		}
	}
	return false
}

// normalizePath returns an absolute, symlink-resolved path when possible
func normalizePath(p string) string {
	p = filepath.Clean(p)
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// copyExecutable copies src to dst atomically with executable permissions
func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err = out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// sameFile reports whether a and b point to the same file
func sameFile(a, b string) bool {
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ai, bi)
}
