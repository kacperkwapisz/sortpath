package updater

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "runtime"
    "path/filepath"
    "strings"
    "time"

    "github.com/kacperkwapisz/sortpath/internal/config"
)

const (
    githubOwner = "kacperkwapisz"
    githubRepo  = "sortpath"
    releaseURL  = "https://api.github.com/repos/%s/%s/releases/latest"
)

type Release struct {
    Version     string
    DownloadURL string
    PublishedAt time.Time
}

type githubRelease struct {
    TagName     string    `json:"tag_name"`
    PublishedAt time.Time `json:"published_at"`
    Assets      []struct {
        Name               string `json:"name"`
        BrowserDownloadURL string `json:"browser_download_url"`
    } `json:"assets"`
}

// GetLastUpdateCheck returns the last time updates were checked
func GetLastUpdateCheck() (time.Time, error) {
    cacheDir := getCacheDir()
    filePath := filepath.Join(cacheDir, "last-check")
    
    data, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return time.Time{}, nil
        }
        return time.Time{}, err
    }
    
    lastCheck, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
    if err != nil {
        return time.Time{}, err
    }
    return lastCheck, nil
}

// SetLastUpdateCheck sets the last time updates were checked
func SetLastUpdateCheck(t time.Time) error {
    cacheDir := getCacheDir()
    if err := os.MkdirAll(cacheDir, 0755); err != nil {
        return err
    }
    
    filePath := filepath.Join(cacheDir, "last-check")
    return os.WriteFile(filePath, []byte(t.Format(time.RFC3339)), 0644)
}

func getCacheDir() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".cache", "sortpath")
}

func CheckLatestRelease() (*Release, error) {
	url := fmt.Sprintf(releaseURL, githubOwner, githubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find appropriate asset for current platform
	platform := runtime.GOOS + "_" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		platform += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, platform) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no suitable binary found for %s", platform)
	}

	return &Release{
		Version:     strings.TrimPrefix(release.TagName, "v"),
		DownloadURL: downloadURL,
		PublishedAt: release.PublishedAt,
	}, nil
}

func UpdateBinary(release *Release) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Download new binary
	resp, err := http.Get(release.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %d", resp.StatusCode)
	}

	// Create temporary file
	tmpPath := execPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpPath) // Clean up on failure

	// Copy new binary
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write update: %w", err)
	}
	tmpFile.Close()

	// Verify the binary is executable
	if err := verifyBinary(tmpPath); err != nil {
		return fmt.Errorf("update verification failed: %w", err)
	}

	// Move temporary file to final location
	if err := os.Rename(tmpPath, execPath); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}

func verifyBinary(path string) error {
	// Simple verification: check if file exists and is executable
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return fmt.Errorf("downloaded binary is empty")
	}
	return nil
}

// IsInstalled returns true if sortpath was installed via the install command
func IsInstalled() bool {
	c, _ := config.Load()
	return c.InstalledPath != ""
}