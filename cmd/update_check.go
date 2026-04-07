package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/loops-so/cli/internal/config"
)

var (
	updateCheckDone   chan struct{}
	updateCheckCancel context.CancelFunc
)

const updateCheckInterval = 24 * time.Hour

type updateCache struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

func checkForUpdate(w io.Writer) {
	if version == "dev" || isJSONOutput() {
		return
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return
	}
	path := filepath.Join(dir, "update-check.json")

	cache, err := readUpdateCache(path)
	if err != nil {
		if debugFlag {
			fmt.Fprintf(w, "[debug] update check: no cache (%v)\n", err)
		}
	} else {
		if debugFlag {
			age := time.Since(cache.CheckedAt).Truncate(time.Second)
			fmt.Fprintf(w, "[debug] update check: cached latest=%s age=%s\n", cache.LatestVersion, age)
		}
		if isNewerVersion(cache.LatestVersion, version) {
			fmt.Fprintf(w, "\nA new version of loops is available: v%s → v%s\nRun this to update:\n\n  %s\n\n", version, cache.LatestVersion, upgradeCommand())
		}
	}

	if err != nil || time.Since(cache.CheckedAt) > updateCheckInterval {
		if debugFlag {
			fmt.Fprintf(w, "[debug] update check: fetching latest release in background\n")
		}
		ctx, cancel := context.WithCancel(context.Background())
		updateCheckCancel = cancel
		updateCheckDone = make(chan struct{})
		go func() {
			defer close(updateCheckDone)
			fetchAndCacheLatestVersion(ctx, path)
		}()
	}
}

func readUpdateCache(path string) (*updateCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c updateCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func fetchAndCacheLatestVersion(ctx context.Context, path string) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/loops-so/cli/releases/latest", nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	if latest == "" {
		return
	}

	cache := updateCache{
		LatestVersion: latest,
		CheckedAt:     time.Now(),
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	_ = os.WriteFile(path, data, 0o600)
}

func upgradeCommand() string {
	if isHomebrew() {
		return "brew upgrade loops"
	}
	installDir := binDir()
	if runtime.GOOS == "windows" {
		if installDir != "" {
			return fmt.Sprintf(`irm https://raw.githubusercontent.com/loops-so/cli/main/install.ps1 | iex -Args "-InstallDir '%s'"`, installDir)
		}
		return `irm https://raw.githubusercontent.com/loops-so/cli/main/install.ps1 | iex`
	}
	if installDir != "" {
		return fmt.Sprintf(`curl -fsSL --proto '=https' --tlsv1.2 https://cli.loops.so | bash -s -- latest %s`, installDir)
	}
	return `curl -fsSL --proto '=https' --tlsv1.2 https://cli.loops.so | bash`
}

func binDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

func isHomebrew() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}
	lower := strings.ToLower(resolved)
	return strings.Contains(lower, "cellar") || strings.Contains(lower, "homebrew") || strings.Contains(lower, "linuxbrew")
}

// isNewerVersion reports whether latest is newer than current (semver without v prefix).
func isNewerVersion(latest, current string) bool {
	l := parseSemver(latest)
	c := parseSemver(current)
	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	var parts [3]int
	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}
