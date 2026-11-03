package linux

import (
	"io"
	"os"
	"regexp"
)

var (
	secretPattern = regexp.MustCompile(
		`(?i)(password|passwd|token|secret|api[_-]?key)\s*[:=]\s*(?:"[^"]*"|'[^']*'|[^\s]+)`,
	)
	privateKeyPattern = regexp.MustCompile(
		`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`,
	)
)

// ReadJournal reads a bounded tail from a regular log file and redacts secrets.
func ReadJournal(path string, maxBytes int64) (string, bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !info.Mode().IsRegular() {
		return "", false, nil
	}
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}
	f, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = f.Close() }()
	start := info.Size() - maxBytes
	if start < 0 {
		start = 0
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return "", false, err
	}
	data, err := io.ReadAll(io.LimitReader(f, maxBytes))
	if err != nil {
		return "", false, err
	}
	return RedactLog(string(data)), true, nil
}

// RedactLog removes common secret assignments and private key blocks.
func RedactLog(text string) string {
	redacted := privateKeyPattern.ReplaceAllString(text, "[REDACTED PRIVATE KEY]")
	return secretPattern.ReplaceAllString(redacted, "$1=[REDACTED]")
}
