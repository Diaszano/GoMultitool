package imageconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Compatibility describes the supported Armbian preset schema and release floor.
type Compatibility struct {
	OSID               string   `json:"os_id"`
	MinimumVersion     string   `json:"minimum_version"`
	SchemaVersion      int      `json:"schema_version"`
	SupportedVariables []string `json:"supported_variables"`
}

// File is an image-relative file artifact with its required mode.
type File struct {
	Data []byte
	Mode os.FileMode
}

// Artifacts contains files to install into an offline root filesystem.
type Artifacts struct {
	FirstBoot []byte
	Files     map[string]File
}

// LoadCompatibility reads the checked-in JSON-compatible YAML profile.
func LoadCompatibility(r io.Reader) (Compatibility, error) {
	var c Compatibility
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	err := d.Decode(&c)
	return c, err
}

// Render creates first-boot, hostname, and optional SSH key files.
func Render(profile Compatibility, version string, p Preset) (Artifacts, error) {
	if profile.OSID != "armbian" || profile.SchemaVersion != 1 ||
		!versionAtLeast(version, profile.MinimumVersion) {
		return Artifacts{}, ErrUnsupportedImage
	}
	if err := p.Validate(); err != nil {
		return Artifacts{}, err
	}
	values := map[string]string{
		"PRESET_ROOT_PASSWORD":        p.RootPassword,
		"PRESET_USER_NAME":            p.UserName,
		"PRESET_USER_PASSWORD":        p.UserPassword,
		"PRESET_DEFAULT_REALNAME":     p.RealName,
		"PRESET_LOCALE":               p.Locale,
		"PRESET_TIMEZONE":             p.Timezone,
		"PRESET_NET_ETHERNET_ENABLED": "1",
		"PRESET_NET_CHANGE_DEFAULTS":  "0",
	}
	for name := range values {
		if !contains(profile.SupportedVariables, name) {
			return Artifacts{}, fmt.Errorf("%w: %s", ErrUnsupportedVariable, name)
		}
	}
	var b bytes.Buffer
	for _, name := range []string{"PRESET_ROOT_PASSWORD", "PRESET_USER_NAME", "PRESET_USER_PASSWORD", "PRESET_DEFAULT_REALNAME", "PRESET_LOCALE", "PRESET_TIMEZONE", "PRESET_NET_ETHERNET_ENABLED", "PRESET_NET_CHANGE_DEFAULTS"} {
		fmt.Fprintf(&b, "%s=\"%s\"\n", name, quote(values[name]))
	}
	files := map[string]File{
		"root/.not_logged_in_yet": {Data: b.Bytes(), Mode: 0o600},
		"etc/hostname":            {Data: []byte(p.Hostname + "\n"), Mode: 0o600},
	}
	if p.SSHAuthorizedKey != "" {
		files["home/"+p.UserName+"/.ssh/authorized_keys"] = File{
			Data: []byte(p.SSHAuthorizedKey + "\n"),
			Mode: 0o600,
		}
	}
	return Artifacts{FirstBoot: b.Bytes(), Files: files}, nil
}

func quote(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "$", "\\$", "`", "\\`").Replace(value)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func versionAtLeast(got, want string) bool {
	g1, g2, ok := parseVersion(got)
	if !ok {
		return false
	}
	w1, w2, ok := parseVersion(want)
	return ok && (g1 > w1 || g1 == w1 && g2 >= w2)
}

func parseVersion(s string) (int, int, bool) {
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}
	major, e1 := strconv.Atoi(parts[0])
	minorPart := strings.TrimLeftFunc(parts[1], func(r rune) bool { return r < '0' || r > '9' })
	minor, e2 := strconv.Atoi(minorPart)
	return major, minor, e1 == nil && e2 == nil && major > 0
}
