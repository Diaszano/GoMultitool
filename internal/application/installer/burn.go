package installer

import (
	"context"
	"errors"
	"io"

	"github.com/Diaszano/GoMultitool/internal/application"
	"github.com/Diaszano/GoMultitool/internal/application/storageguard"
	"github.com/Diaszano/GoMultitool/internal/domain/installation"
	"github.com/Diaszano/GoMultitool/internal/platform/rk322x"
)

// Burn writes an image and, on rk322x, carries the DDR command rate over.
type Burn struct {
	Guard      storageguard.Guard
	OpenTarget func(string) (application.Target, error)
	Platform   string // "rk322x" enables command-rate preservation
}

// BurnResult reports non-fatal outcomes of a successful burn.
type BurnResult struct{ CommandRateNotApplied bool }

// Run validates, writes and verifies the image, then restores the command rate.
func (b Burn) Run(
	ctx context.Context,
	plan installation.Plan,
	source io.Reader,
) (BurnResult, error) {
	if err := plan.Validate(); err != nil {
		return BurnResult{}, err
	}
	rk := b.Platform == "rk322x" && b.OpenTarget != nil
	var saved byte
	var found bool
	if rk {
		_, rate, err := b.findRate(plan.Target.Path)
		if err != nil && !errors.Is(err, rk322x.ErrNoCommandRate) {
			return BurnResult{}, err
		}
		saved, found = rate, err == nil
	}
	if err := (Writer{}).Execute(ctx, plan, source, true, b.Guard, b.OpenTarget); err != nil {
		return BurnResult{}, err
	}
	if !found {
		return BurnResult{}, nil
	}
	return b.restoreRate(plan.Target.Path, saved)
}

// findRate opens the target, scans its head and always closes it.
func (b Burn) findRate(path string) (offset int64, rate byte, retErr error) {
	t, err := b.OpenTarget(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = t.Close() }()
	head := make([]byte, rk322x.CommandRateScan)
	n, err := io.ReadFull(t, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return 0, 0, err
	}
	return rk322x.FindCommandRate(head[:n])
}

func (b Burn) restoreRate(path string, saved byte) (res BurnResult, retErr error) {
	t, err := b.OpenTarget(path)
	if err != nil {
		return BurnResult{}, err
	}
	defer func() {
		if err := t.Close(); retErr == nil {
			retErr = err
		}
	}()
	head := make([]byte, rk322x.CommandRateScan)
	n, err := io.ReadFull(t, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return BurnResult{}, err
	}
	offset, _, err := rk322x.FindCommandRate(head[:n])
	if err != nil {
		return BurnResult{CommandRateNotApplied: true}, nil
	}
	if _, err := t.Seek(offset, io.SeekStart); err != nil {
		return BurnResult{}, err
	}
	if _, err := t.Write([]byte{saved}); err != nil {
		return BurnResult{}, err
	}
	return BurnResult{}, t.Sync()
}
