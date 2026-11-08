package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/backup"
	"github.com/Diaszano/GoMultitool/internal/application/discovery"
	"github.com/Diaszano/GoMultitool/internal/application/installer"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/device"
	"github.com/Diaszano/GoMultitool/internal/domain/image"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
	"github.com/Diaszano/GoMultitool/internal/infrastructure/compression"
	linuxsys "github.com/Diaszano/GoMultitool/internal/infrastructure/linux"
	"github.com/Diaszano/GoMultitool/internal/infrastructure/multitool"
	"github.com/Diaszano/GoMultitool/internal/presentation/tui"
)

const lockPath = "/run/multitool.lock"

// openImage is a seam so tests can fake the decompressor.
var openImage = compression.Open

// services wires the real Linux adapters into the TUI.
func services(catalog linuxsys.SysfsCatalog, locator multitool.Locator) tui.Services {
	return tui.Services{
		Inventory: discovery.Service{Catalog: catalog}.Inventory,
		Images: func(ctx context.Context) (images []image.Image, err error) {
			err = withWorkspace(ctx, locator, func(ws multitool.Workspace) error {
				images, err = compression.ListImages(ctx, ws.ImagesDir())
				return err
			})
			return images, err
		},
		FreeBytes: func(ctx context.Context) (free int64, err error) {
			err = withWorkspace(ctx, locator, func(ws multitool.Workspace) error {
				free, err = ws.FreeBytes()
				return err
			})
			return free, err
		},
		Burn: func(
			ctx context.Context, img image.Image, target device.Device, p tui.Progress,
		) (result installer.BurnResult, err error) {
			err = locked(func() error {
				return withWorkspace(ctx, locator, func(ws multitool.Workspace) error {
					result, err = burn(ctx, catalog, ws, img, target, p)
					return err
				})
			})
			return result, err
		},
		Backup: func(ctx context.Context, source device.Device, name string, p tui.Progress) error {
			return locked(func() error {
				return withWorkspace(ctx, locator, func(ws multitool.Workspace) error {
					return backupTo(ctx, ws, source, name, p)
				})
			})
		},
		RunLegacy: func() error {
			cmd := exec.Command("/usr/local/bin/multitool.sh")
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			return cmd.Run()
		},
	}
}

func burn(
	ctx context.Context,
	catalog application.DeviceCatalog,
	ws multitool.Workspace,
	img image.Image,
	target device.Device,
	p tui.Progress,
) (_ installer.BurnResult, err error) {
	if img.SizeBytes == 0 {
		inspected, err := compression.Inspect(ctx, img.Path)
		if err != nil {
			return installer.BurnResult{}, err
		}
		img = inspected
	}
	plan := installation.Plan{
		Image:   img,
		Target:  target,
		Summary: fmt.Sprintf("%s -> %s", filepath.Base(img.Path), target.Path),
	}
	src, err := openImage(ctx, img.Path)
	if err != nil {
		return installer.BurnResult{}, err
	}
	// For .img.xz, Close is the only place xz's exit status (integrity check) shows up.
	defer func() { err = errors.Join(err, src.Close()) }()
	return installer.Burn{
		Guard: storageguard.Guard{Catalog: catalog},
		OpenTarget: func(path string) (application.Target, error) {
			return os.OpenFile(path, os.O_RDWR, 0)
		},
		Platform: ws.Platform,
	}.Run(ctx, plan, &countingReader{r: src, total: img.SizeBytes, p: p})
}

func backupTo(
	ctx context.Context,
	ws multitool.Workspace,
	source device.Device,
	name string,
	p tui.Progress,
) error {
	destination, err := backupDestination(ws.BackupsDir(), name)
	if err != nil {
		return err
	}
	// Read-only: a backup never needs to write to its source.
	src, err := os.Open(source.Path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	_, err = backup.Service{}.Create(
		ctx, source, &countingReader{r: src, total: source.SizeBytes, p: p}, destination)
	return err
}

// backupDestination keeps the backup inside dir whatever the typed name is.
func backupDestination(dir, name string) (string, error) {
	if name == "" || strings.HasPrefix(name, ".") || strings.ContainsRune(name, '/') {
		return "", fmt.Errorf("invalid backup name %q", name)
	}
	return filepath.Join(dir, name+".img.gz"), nil
}

func locked(op func() error) (err error) {
	unlock, err := multitool.Lock(lockPath)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, unlock()) }()
	return op()
}

// withWorkspace opens MULTITOOL for one operation; a failed unmount fails the operation.
func withWorkspace(
	ctx context.Context,
	locator multitool.Locator,
	op func(multitool.Workspace) error,
) (err error) {
	ws, closeWorkspace, err := locator.Open(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeWorkspace()) }()
	return op(ws)
}

// countingReader reports the bytes read so far to p, at most once per percent or 200 ms,
// plus once when the total is reached or the source ends.
type countingReader struct {
	r        io.Reader
	done     int64
	total    int64
	p        tui.Progress
	reported int64
	percent  int64
	at       time.Time
}

func (c *countingReader) Read(b []byte) (int, error) {
	n, err := c.r.Read(b)
	c.done += int64(n)
	if c.p == nil || c.done == c.reported {
		return n, err
	}
	percent := int64(-1)
	if c.total > 0 {
		percent = c.done * 100 / c.total
	}
	final := err != nil || (c.total > 0 && c.reported < c.total && c.done >= c.total)
	if final || percent != c.percent || time.Since(c.at) >= 200*time.Millisecond {
		c.p(c.done, c.total)
		c.reported, c.percent, c.at = c.done, percent, time.Now()
	}
	return n, err
}
