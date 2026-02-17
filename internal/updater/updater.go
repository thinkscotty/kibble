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

// RestartService attempts to restart the kibble systemd service.
func RestartService() error {
	if path, err := exec.LookPath("systemctl"); err == nil {
		slog.Info("Restarting via systemctl", "service", "kibble")
		cmd := exec.Command(path, "restart", "kibble")
		output, err := cmd.CombinedOutput()
		if err != nil {
			slog.Warn("systemctl restart failed, falling back to process exit",
				"error", err, "output", string(output))
		} else {
			slog.Info("systemctl restart succeeded")
			return nil
		}
	} else {
		slog.Info("systemctl not found, using process exit for restart")
	}

	// Fallback: exit cleanly and let systemd restart us
	// This works with Restart=always in the systemd service
	slog.Info("Exiting process to trigger systemd restart")
	time.Sleep(100 * time.Millisecond) // Brief delay to flush logs
	os.Exit(0)
	return nil
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
