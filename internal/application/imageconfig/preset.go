// Package imageconfig renders validated offline Linux image configuration.
package imageconfig

import (
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidPreset indicates missing or unsafe user-supplied settings.
	ErrInvalidPreset = errors.New("invalid provisioning preset")
	// ErrUnsupportedImage indicates an OS ID or version outside the profile.
	ErrUnsupportedImage = errors.New("unsupported Armbian image")
	// ErrUnsupportedVariable indicates the image profile lacks a required directive.
	ErrUnsupportedVariable = errors.New("unsupported Armbian first-boot variable")
)

var (
	userPattern   = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
	hostPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,62}$`)
	localePattern = regexp.MustCompile(`^[A-Za-z]{2,3}_[A-Za-z]{2,3}\.UTF-8$`)
	zonePattern   = regexp.MustCompile(`^[A-Za-z0-9_+-]+(/[A-Za-z0-9_+-]+)*$`)
)

// Preset contains requested first-boot settings. Passwords and keys are never logged.
type Preset struct {
	RootPassword     string
	UserName         string
	UserPassword     string
	RealName         string
	Hostname         string
	Locale           string
	Timezone         string
	SSHAuthorizedKey string
}

// Validate rejects unsafe identifiers, secrets, and public-key values.
func (p Preset) Validate() error {
	if !userPattern.MatchString(p.UserName) || len(p.RootPassword) < 8 ||
		len(p.RootPassword) > 128 ||
		len(p.UserPassword) < 8 ||
		len(p.UserPassword) > 128 ||
		p.RealName == "" ||
		strings.ContainsAny(p.RealName, "\r\n\x00") ||
		!hostPattern.MatchString(p.Hostname) ||
		!localePattern.MatchString(p.Locale) ||
		!zonePattern.MatchString(p.Timezone) ||
		strings.ContainsAny(p.RootPassword+p.UserPassword, "\r\n\x00") {
		return ErrInvalidPreset
	}
	if p.SSHAuthorizedKey != "" {
		fields := strings.Fields(p.SSHAuthorizedKey)
		if len(fields) < 2 || (fields[0] != "ssh-ed25519" && fields[0] != "ssh-rsa") {
			return ErrInvalidPreset
		}
		key, err := base64.StdEncoding.DecodeString(fields[1])
		if err != nil || len(key) == 0 || strings.ContainsAny(p.SSHAuthorizedKey, "\r\n\x00") {
			return ErrInvalidPreset
		}
	}
	return nil
}
