package browser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	tagUrl = `https://api.github.com/repos/ValentinGrim/final-fantasy-pr-save-editor/tags`
	relUrl = `https://github.com/ValentinGrim/final-fantasy-pr-save-editor/releases/%s`
)

// Version is overridden at build time from the git tag:
// -ldflags "-X pixel-remastered-save-editor/browser.Version=1.4.0"
var Version = "1.4.0"

type (
	tag struct {
		Name string `json:"name"`
	}
	comp struct {
		version [3]int
		// pre is set for a prerelease ("1.5.0-rc1"), which comes before the release with the same numbers.
		pre bool
	}
)

var client = &http.Client{Timeout: 10 * time.Second}

// CheckForUpdate looks for a release tag newer than Version. Prerelease tags are never offered.
// It runs on a background goroutine, so it must never panic: a panic there would kill the app.
func CheckForUpdate() (hasNewer bool, version string, err error) {
	defer func() {
		if r := recover(); r != nil {
			hasNewer, version, err = false, "", fmt.Errorf("update check failed: %v", r)
		}
	}()
	var (
		r          *http.Response
		b          []byte
		tags       []tag
		current, _ = newComparable(Version)
	)
	if r, err = client.Get(tagUrl); err != nil {
		return
	}
	defer func() { _ = r.Body.Close() }()
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("update check failed: %s", r.Status)
		return
	}
	if b, err = io.ReadAll(io.LimitReader(r.Body, 1<<20)); err != nil {
		return
	}
	if err = json.Unmarshal(b, &tags); err != nil {
		return
	}

	for _, t := range tags {
		if other, ok := newComparable(t.Name); ok && !other.pre {
			if IsNewer(current, other) {
				current = other
				hasNewer = true
				version = t.Name
			}
		}
	}
	return
}

// newComparable parses "[v]MAJOR.MINOR[.PATCH][-PRERELEASE][+BUILD]". Tags containing "_", with more than three
// numbers or with anything else than digits in a number are not comparable.
func newComparable(v string) (c comp, found bool) {
	v = strings.TrimSpace(v)
	if strings.Contains(v, "_") {
		return
	}
	v = strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		c.pre = true
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > len(c.version) {
		return comp{}, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil { // signs were cut off above, so only plain digits get here
			return comp{}, false
		}
		c.version[i] = n
	}
	found = true
	return
}

// IsNewer reports whether other is a later version than current.
func IsNewer(current, other comp) bool {
	for i := 0; i < len(current.version); i++ {
		c := current.version[i]
		o := other.version[i]
		if c < o {
			return true
		} else if c != o {
			return false
		}
	}
	return current.pre && !other.pre
}

// Update shows the release page of tag in the default browser.
func Update(tag string) error {
	return start(fmt.Sprintf(relUrl, tag), "rundll32", "url.dll,FileProtocolHandler")
}

// OpenFolder shows dir in the system file manager.
func OpenFolder(dir string) error {
	return start(dir, "explorer")
}

// start hands target to the desktop: xdg-open on Linux, open on macOS and the
// windows command line (target appended) on Windows. It does not wait for the
// handler.
func start(target string, windows ...string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", target)
	case "windows":
		cmd = exec.Command(windows[0], append(windows[1:], target)...)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		return fmt.Errorf("unsupported platform")
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	// Reap the process once the handler exits.
	go func() { _ = cmd.Wait() }()
	return nil
}
