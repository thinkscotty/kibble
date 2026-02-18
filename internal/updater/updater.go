package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ReleaseInfo holds the result of a version check against GitHub.
type ReleaseInfo struct {
	TagName     string
	Version     string // TagName with leading "v" stripped
	PublishedAt string
	HTMLURL     string
	Body        string // release notes (markdown)
	AssetURL    string // direct download URL for the correct binary
	AssetName   string
	AssetSize   int64
}

// UpdateResult describes what happened during an install attempt.
type UpdateResult struct {
	OldVersion string
	NewVersion string
	AssetName  string
}

// GitHub API response types.
type ghRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt string    `json:"published_at"`
	Assets      []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var installMu sync.Mutex

const githubAPI = "https://api.github.com/repos/thinkscotty/kibble/releases/latest"

// CheckForUpdate queries the GitHub releases API and returns info about the
// latest release, or nil if the current version is already up-to-date.
func CheckForUpdate(ctx context.Context, currentVersion string) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "kibble-updater")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parse GitHub response: %w", err)
	}

	asset, ok := matchAsset(release.Assets)
	if !ok {
		return nil, fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if !isNewer(currentVersion, latestVersion) {
		return nil, nil // already up-to-date
	}

	return &ReleaseInfo{
		TagName:     release.TagName,
		Version:     latestVersion,
		PublishedAt: release.PublishedAt,
		HTMLURL:     release.HTMLURL,
		Body:        release.Body,
		AssetURL:    asset.BrowserDownloadURL,
		AssetName:   asset.Name,
		AssetSize:   asset.Size,
	}, nil
}

// DownloadAndInstall downloads the release asset and atomically replaces the
// running binary.
func DownloadAndInstall(ctx context.Context, info *ReleaseInfo, currentVersion string) (*UpdateResult, error) {
	if !installMu.TryLock() {
		return nil, fmt.Errorf("an update is already in progress")
	}
	defer installMu.Unlock()

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("find executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("resolve symlinks: %w", err)
	}

	dir := filepath.Dir(execPath)
	if err := checkWritable(dir); err != nil {
		return nil, fmt.Errorf("cannot update: no write permission to %s", dir)
	}

	tmpPath := execPath + ".update.tmp"
	os.Remove(tmpPath) // clean up any stale temp file

	slog.Info("Downloading update", "url", info.AssetURL, "target", tmpPath)

	req, err := http.NewRequestWithContext(ctx, "GET", info.AssetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("User-Agent", "kibble-updater")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	written, copyErr := io.Copy(f, resp.Body)
	f.Close()

	if copyErr != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("write download: %w", copyErr)
	}

	if info.AssetSize > 0 && written != info.AssetSize {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("download size mismatch: expected %d bytes, got %d", info.AssetSize, written)
	}

	slog.Info("Download complete", "bytes", written)

	// Preserve SELinux context if applicable
	preserveSELinuxContext(execPath, tmpPath)

	// Atomic replace
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("replace binary: %w", err)
	}

	slog.Info("Binary replaced successfully", "path", execPath)

	return &UpdateResult{
		OldVersion: currentVersion,
		NewVersion: info.Version,
		AssetName:  info.AssetName,
	}, nil
}

// RestartService attempts to restart the service using multiple strategies.
// It tries them in order until one succeeds:
//  1. systemctl restart <detected-unit> (auto-detected from /proc/self/cgroup)
//  2. systemctl restart kibble (well-known service name)
//  3. service kibble restart (SysV init compatibility)
//  4. Self-exec: replace the current process with the new binary (works without any service manager)
//  5. os.Exit(0): exit and rely on the service manager's Restart=always directive
func RestartService() error {
	// Strategy 1: Detect the systemd unit managing this process and restart it
	if unit := detectSystemdUnit(); unit != "" {
		slog.Info("Detected systemd unit", "unit", unit)
		if trySystemctl("restart", unit) {
			return nil
		}
	}

	// Strategy 2: Try well-known service name
	if trySystemctl("restart", "kibble") {
		return nil
	}

	// Strategy 3: SysV init compatibility (Debian/Ubuntu legacy, some containers)
	if path, err := exec.LookPath("service"); err == nil {
		slog.Info("Trying SysV service restart")
		cmd := exec.Command(path, "kibble", "restart")
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("SysV service restart failed", "error", err, "output", string(output))
		} else {
			slog.Info("SysV service restart succeeded")
			return nil
		}
	}

	// Strategy 4: Self-exec — replace this process with the new binary.
	// This works regardless of init system. The new binary starts fresh with
	// the same PID, arguments, and environment. syscall.Exec is Unix-only;
	// on Windows it returns an error and we fall through.
	execPath, err := os.Executable()
	if err == nil {
		execPath, _ = filepath.EvalSymlinks(execPath)
		slog.Info("Attempting self-exec restart", "binary", execPath, "args", os.Args)
		// syscall.Exec replaces the process image — if it succeeds, this line
		// is the last thing the old process ever runs. The new binary starts
		// from main() with the same PID.
		if err := syscall.Exec(execPath, os.Args, os.Environ()); err != nil {
			slog.Warn("Self-exec failed", "error", err)
		}
	}

	// Strategy 5: Exit cleanly and hope the service manager restarts us.
	// Works with systemd Restart=always, Docker restart policies, etc.
	slog.Info("Exiting process for service manager restart")
	time.Sleep(100 * time.Millisecond) // flush logs
	os.Exit(0)
	return nil
}

// trySystemctl attempts to run "systemctl <action> <unit>" and returns true on success.
func trySystemctl(action, unit string) bool {
	// Try PATH lookup first, then fall back to common absolute paths.
	// systemd services often have a restricted PATH that may not include /usr/bin.
	path, err := exec.LookPath("systemctl")
	if err != nil {
		for _, p := range []string{"/usr/bin/systemctl", "/bin/systemctl"} {
			if _, serr := os.Stat(p); serr == nil {
				path = p
				break
			}
		}
	}
	if path == "" {
		return false
	}

	slog.Info("Trying systemctl restart", "path", path, "unit", unit)
	cmd := exec.Command(path, action, unit)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("systemctl restart failed", "unit", unit, "error", err, "output", string(output))
		return false
	}
	slog.Info("systemctl restart succeeded", "unit", unit)
	return true
}

// detectSystemdUnit reads /proc/self/cgroup to find the systemd service unit
// managing this process. Returns the unit name (e.g. "kibble.service") or "".
func detectSystemdUnit() string {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "" // not Linux, or no cgroup info
	}
	for _, line := range strings.Split(string(data), "\n") {
		// cgroup v2: "0::/system.slice/kibble.service"
		// cgroup v1: "1:name=systemd:/system.slice/kibble.service"
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		cgpath := parts[2]
		segments := strings.Split(cgpath, "/")
		for _, seg := range segments {
			if strings.HasSuffix(seg, ".service") {
				return seg
			}
		}
	}
	return ""
}

// matchAsset finds the GitHub release asset matching the current platform.
func matchAsset(assets []ghAsset) (ghAsset, bool) {
	wantName := fmt.Sprintf("kibble-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		wantName += ".exe"
	}
	for _, a := range assets {
		if a.Name == wantName {
			return a, true
		}
	}
	return ghAsset{}, false
}

// isNewer returns true if latest is newer than current.
func isNewer(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Strip -dirty suffix
	current = strings.TrimSuffix(current, "-dirty")

	// Handle git describe output like "0.8.2-3-gabcdef1"
	// Extract just the version number before the dash
	if idx := strings.Index(current, "-"); idx > 0 {
		current = current[:idx]
	}

	// If current is "dev" or a commit hash, any release is newer
	if current == "dev" || current == "unknown" || !isSemver(current) {
		return true
	}

	curParts := parseSemver(current)
	latParts := parseSemver(latest)

	for i := 0; i < 3; i++ {
		if latParts[i] > curParts[i] {
			return true
		}
		if latParts[i] < curParts[i] {
			return false
		}
	}
	return false // equal
}

func isSemver(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

func parseSemver(s string) [3]int {
	var result [3]int
	parts := strings.Split(s, ".")
	for i := 0; i < len(parts) && i < 3; i++ {
		result[i], _ = strconv.Atoi(parts[i])
	}
	return result
}

// checkWritable tests if the directory is writable by creating a temp file.
func checkWritable(dir string) error {
	testPath := filepath.Join(dir, ".kibble-write-test")
	f, err := os.Create(testPath)
	if err != nil {
		return err
	}
	f.Close()
	os.Remove(testPath)
	return nil
}

// preserveSELinuxContext copies the SELinux security context from src to dst.
// On systems without SELinux this is a silent no-op.
func preserveSELinuxContext(src, dst string) {
	getfattr, err := exec.LookPath("getfattr")
	if err != nil {
		return
	}
	setfattr, err := exec.LookPath("setfattr")
	if err != nil {
		return
	}

	out, err := exec.Command(getfattr, "--name=security.selinux", "--only-values", src).Output()
	if err != nil {
		slog.Debug("Could not read SELinux context", "error", err)
		return
	}

	ctx := strings.TrimSpace(string(out))
	if ctx == "" {
		return
	}

	if err := exec.Command(setfattr, "--name=security.selinux", "--value="+ctx, dst).Run(); err != nil {
		slog.Warn("Could not set SELinux context on new binary", "error", err)
	} else {
		slog.Debug("Preserved SELinux context", "context", ctx)
	}
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
