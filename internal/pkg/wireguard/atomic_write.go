package wireguard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// AtomicWriteFile writes content to path atomically (write temp file then rename).
// If the target file exists, it will create a timestamped backup next to it.
func AtomicWriteFile(path string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Backup existing file if it exists.
	if st, err := os.Stat(path); err == nil && st.Mode().IsRegular() {
		ts := time.Now().Format("20060102-150405")
		bak := filepath.Join(dir, fmt.Sprintf("%s.bak.%s", base, ts))
		if err := copyFile(path, bak, st.Mode().Perm()); err != nil {
			return err
		}
		// Preserve existing perms unless caller explicitly sets.
		if perm == 0 {
			perm = st.Mode().Perm()
		}
	}

	if perm == 0 {
		perm = 0600
	}

	tmp, err := os.CreateTemp(dir, base+".tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// Rename is atomic on the same filesystem.
	return os.Rename(tmpName, path)
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
