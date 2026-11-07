// Package multitool locates the MULTITOOL partition and serializes operations.
package multitool

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Diaszano/GoMultitool/internal/infrastructure/filesystem"
)

// ErrBusy means another operation holds the lock.
var ErrBusy = errors.New("another operation is running")

// Locator finds (and mounts when needed) the MULTITOOL partition.
type Locator struct {
	Runner     filesystem.CommandRunner // nil runs real commands
	ProcRoot   string                   // default "/proc"
	MountPoint string                   // default "/mnt"
}

// Workspace is a mounted MULTITOOL partition.
type Workspace struct {
	Root     string
	Platform string
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// Open returns the workspace and a close func that unmounts only what Open mounted.
func (l Locator) Open(ctx context.Context) (Workspace, func() error, error) {
	runner, proc, mp := l.Runner, l.ProcRoot, l.MountPoint
	if runner == nil {
		runner = execRunner{}
	}
	if proc == "" {
		proc = "/proc"
	}
	if mp == "" {
		mp = "/mnt"
	}
	out, err := runner.Run(ctx, "blkid", "-l", "-o", "device", "--label", "MULTITOOL")
	dev := strings.TrimSpace(string(out))
	if dev == "" {
		return Workspace{}, nil, errors.New("MULTITOOL partition not found")
	}
	if err != nil {
		return Workspace{}, nil, fmt.Errorf("blkid: %w", err)
	}
	root, closeFn := mountedAt(proc, dev), func() error { return nil }
	if root == "" {
		if out, err := runner.Run(ctx, "mount", dev, mp); err != nil {
			return Workspace{}, nil, fmt.Errorf(
				"mount MULTITOOL: %w: %s", err, strings.TrimSpace(string(out)))
		}
		root = mp
		closeFn = func() error {
			if out, err := runner.Run(context.Background(), "umount", mp); err != nil {
				return fmt.Errorf("umount MULTITOOL: %w: %s", err, strings.TrimSpace(string(out)))
			}
			return nil
		}
	}
	ws := Workspace{Root: root}
	if b, err := os.ReadFile(filepath.Join(root, "TARGET")); err == nil {
		ws.Platform = strings.TrimSpace(string(b))
	}
	if err := os.MkdirAll(ws.BackupsDir(), 0o755); err != nil {
		_ = closeFn()
		return Workspace{}, nil, err
	}
	return ws, closeFn, nil
}

func mountedAt(procRoot, dev string) string {
	f, err := os.Open(filepath.Join(procRoot, "mounts"))
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if fields := strings.Fields(s.Text()); len(fields) >= 2 && fields[0] == dev {
			return fields[1]
		}
	}
	return ""
}

// ImagesDir is where burnable images live.
func (w Workspace) ImagesDir() string { return filepath.Join(w.Root, "images") }

// BackupsDir is where backups are written.
func (w Workspace) BackupsDir() string { return filepath.Join(w.Root, "backups") }

// FreeBytes reports space available to unprivileged writers on the partition.
func (w Workspace) FreeBytes() (int64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(w.Root, &st); err != nil {
		return 0, err
	}
	return int64(st.Bavail) * int64(st.Bsize), nil
}

// Lock takes the single-operation lock; the returned func releases it.
func Lock(path string) (func() error, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrBusy
		}
		return nil, err
	}
	return f.Close, nil // closing drops the flock
}
