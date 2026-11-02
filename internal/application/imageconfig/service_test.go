package imageconfig

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeMounter struct {
	root    string
	fail    error
	cleaned bool
	cancel  context.CancelFunc
}

func (m *fakeMounter) Mount(context.Context, string) (string, func() error, error) {
	if m.fail != nil {
		return "", nil, m.fail
	}
	if m.cancel != nil {
		m.cancel()
	}
	return m.root, func() error { m.cleaned = true; return nil }, nil
}
func supportedPreset() Preset {
	return Preset{RootPassword: "RootSecret123", UserName: "armbian", UserPassword: "UserSecret123", RealName: "Armbian User", Hostname: "rk3229", Locale: "en_US.UTF-8", Timezone: "Etc/UTC"}
}
func supportedProfile() Compatibility {
	return Compatibility{OSID: "armbian", MinimumVersion: "26.08", SchemaVersion: 1, SupportedVariables: []string{"PRESET_ROOT_PASSWORD", "PRESET_USER_NAME", "PRESET_USER_PASSWORD", "PRESET_DEFAULT_REALNAME", "PRESET_LOCALE", "PRESET_TIMEZONE", "PRESET_NET_ETHERNET_ENABLED", "PRESET_NET_CHANGE_DEFAULTS"}}
}
func sourceImage(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.img")
	if err := os.WriteFile(path, []byte("image-bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPrepareCommitsOnlyAfterFilesAndCleanup(t *testing.T) {
	root := t.TempDir()
	m := &fakeMounter{root: root}
	source := sourceImage(t)
	out := filepath.Join(t.TempDir(), "ready.img")
	if err := (Service{Profile: supportedProfile(), Mounter: m}).Prepare(context.Background(), source, out, "26.08", supportedPreset()); err != nil {
		t.Fatal(err)
	}
	if !m.cleaned {
		t.Fatal("mount was not cleaned")
	}
	data, err := os.ReadFile(out)
	if err != nil || string(data) != "image-bytes" {
		t.Fatalf("image=%q err=%v", data, err)
	}
	info, err := os.Stat(filepath.Join(root, "root/.not_logged_in_yet"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
}
func TestPrepareFailuresLeaveSourceAndNoReadyOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		profile  Compatibility
		mountErr error
	}{{"unsupported", Compatibility{OSID: "debian"}, nil}, {"mount", supportedProfile(), errors.New("mount failed")}} {
		t.Run(tc.name, func(t *testing.T) {
			m := &fakeMounter{root: t.TempDir(), fail: tc.mountErr}
			src := sourceImage(t)
			original, _ := os.ReadFile(src)
			out := filepath.Join(t.TempDir(), "ready.img")
			err := (Service{Profile: tc.profile, Mounter: m}).Prepare(context.Background(), src, out, "26.08", supportedPreset())
			if err == nil {
				t.Fatal("Prepare succeeded")
			}
			got, _ := os.ReadFile(src)
			if string(got) != string(original) {
				t.Fatal("source changed")
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("output exists: %v", err)
			}
		})
	}
}
func TestPrepareWriteFailureCleansMount(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "etc"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	m := &fakeMounter{root: root}
	out := filepath.Join(t.TempDir(), "ready.img")
	err := (Service{Profile: supportedProfile(), Mounter: m}).Prepare(context.Background(), sourceImage(t), out, "26.08", supportedPreset())
	if err == nil {
		t.Fatal("write failure ignored")
	}
	if !m.cleaned {
		t.Fatal("mount not cleaned")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("output exists: %v", err)
	}
}
func TestPrepareRejectsSymlinkedImageSource(t *testing.T) {
	target := sourceImage(t)
	link := filepath.Join(t.TempDir(), "link.img")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	m := &fakeMounter{root: t.TempDir()}
	err := (Service{Profile: supportedProfile(), Mounter: m}).Prepare(context.Background(), link, filepath.Join(t.TempDir(), "out.img"), "26.08", supportedPreset())
	if err == nil {
		t.Fatal("symlink source accepted")
	}
}

func TestPrepareCancellationCleansMountAndDoesNotPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	root := t.TempDir()
	m := &fakeMounter{root: root, cancel: cancel}
	out := filepath.Join(t.TempDir(), "ready.img")
	err := (Service{Profile: supportedProfile(), Mounter: m}).Prepare(ctx, sourceImage(t), out, "26.08", supportedPreset())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	if !m.cleaned {
		t.Fatal("mount not cleaned")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("output exists: %v", err)
	}
}
