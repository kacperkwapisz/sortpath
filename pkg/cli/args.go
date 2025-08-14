package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"errors"
	"time"

	"github.com/kacperkwapisz/sortpath/internal/config"
    "github.com/kacperkwapisz/sortpath/internal/updater"
)

type CLIOptions struct {
	APIKey  string
	APIBase string
	Model   string
	Tree    string
}

func ParseArgs(args []string) (CLIOptions, string) {
	var opts CLIOptions
	fs := flag.NewFlagSet("sortpath", flag.ContinueOnError)
	fs.StringVar(&opts.APIKey, "api-key", "", "OpenAI-compatible API key")
	fs.StringVar(&opts.APIBase, "api-base", "", "API base URL")
	fs.StringVar(&opts.Model, "model", "", "Model name")
	fs.StringVar(&opts.Tree, "tree", "", "Path to folder tree file")
	fs.SetOutput(os.Stderr)

	// Find first non-flag arg as description
	descIdx := 0
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			descIdx = i
			break
		}
	}
	flagArgs := args[:descIdx]
	desc := strings.Join(args[descIdx:], " ")

	_ = fs.Parse(flagArgs)
	return opts, desc
}

func PrintHelp(version string) {
	fmt.Printf(`sortpath: AI-powered folder recommendation CLI
Version: %s

Usage:
  sortpath [flags] "file description"
  sortpath config set|get|remove|list [key] [value]
  sortpath install [--path /usr/local/bin] [--force]
	sortpath update [--check-only]

Flags:
  --api-key   OpenAI-compatible API key
  --api-base  API base URL (e.g. https://api.openai.com/v1)
  --model     Model name (e.g. gpt-3.5-turbo)
  --tree      Path to folder tree file
	-v, --version  Show version

Config subcommands:
  config set <key> <value>
  config get <key>
  config remove <key>
  config list

Install:
  install           Install the current binary to a PATH directory (default /usr/local/bin)
  Options:
    --path PATH     Destination directory (must be on your PATH)
    --force         Overwrite existing binary if present

Update:
	update            Update to the latest version from GitHub
	Options:
	--check-only    Only check for updates, don't install
`, version)
}

func HandleConfigCommand(args []string) {
	if len(args) < 1 {
		PrintHelp("dev") // REPOMARK:SCOPE: 1 - Add version parameter to PrintHelp call
		return
	}
	switch args[0] {
	case "set":
		if len(args) != 3 {
			fmt.Println("Usage: sortpath config set <key> <value>")
			return
		}
		err := config.Set(args[1], args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config set error: %v\n", err)
			os.Exit(1)
		}
	case "get":
		if len(args) != 2 {
			fmt.Println("Usage: sortpath config get <key>")
			return
		}
		val, err := config.Get(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config get error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(val)
	case "remove":
		if len(args) != 2 {
			fmt.Println("Usage: sortpath config remove <key>")
			return
		}
		err := config.Remove(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config remove error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Config list error: %v\n", err)
			os.Exit(1)
		}
		for k, v := range conf.ToMap() {
			fmt.Printf("%s: %s\n", k, v)
		}
	default:
		PrintHelp("dev") // REPOMARK:SCOPE: 2 - Add version parameter to PrintHelp call
	}
}

func HandleInstallCommand(args []string) {
	var destDir string
	var force bool
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.StringVar(&destDir, "path", "/usr/local/bin", "Destination directory (must be on PATH)")
	fs.BoolVar(&force, "force", false, "Overwrite existing binary if present")
	fs.SetOutput(os.Stderr)
	_ = fs.Parse(args)

	srcPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine current executable path: %v\n", err)
		os.Exit(1)
	}

	destPath := filepath.Join(destDir, "sortpath")
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			fmt.Fprintf(os.Stderr, "Destination already has sortpath: %s (use --force to overwrite)\n", destPath)
			os.Exit(1)
		}
	}

	if err := copyFile(srcPath, destPath); err != nil {
		// Permission denied -> fallback to user bin
		if errors.Is(err, os.ErrPermission) || strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			fallbackDir := userBinFallbackDir()
			if fallbackDir == "" {
				fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Try: sudo cp %q %q\n", srcPath, destPath)
				os.Exit(1)
			}
			_ = os.MkdirAll(fallbackDir, 0755)
			userDest := filepath.Join(fallbackDir, "sortpath")
			if err2 := copyFile(srcPath, userDest); err2 != nil {
				fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Also failed to install to %s: %v\n", userDest, err2)
				fmt.Fprintf(os.Stderr, "Try: sudo cp %q %q\n", srcPath, destPath)
				os.Exit(1)
			}
			_ = os.Chmod(userDest, 0755)
			_ = config.Set("installed-path", userDest)

			// Ensure PATH contains fallbackDir; if not, attempt to add to shell profile
			if !pathContainsDir(fallbackDir) {
				profilePath, added, addErr := addDirToShellPATH(fallbackDir)
				if addErr == nil && added {
					fmt.Printf("Installed sortpath to %s and added it to PATH in %s. Restart your shell or run: source %s\n", userDest, profilePath, profilePath)
				} else {
					fmt.Printf("Installed sortpath to %s. Add it to your PATH by adding this to your shell profile:\n\n    export PATH=\"%s:$PATH\"\n\nThen restart your terminal.\n", userDest, fallbackDir)
				}
			} else {
				fmt.Printf("Installed sortpath to %s\n", userDest)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Try: sudo cp %q %q\n", srcPath, destPath)
		os.Exit(1)
	}
	// Make executable
	_ = os.Chmod(destPath, 0755)

	// Save installed path
	_ = config.Set("installed-path", destPath)
	fmt.Printf("Installed sortpath to %s\n", destPath)
}

func HandleUpdateCommand(args []string, currentVersion string) {
    var checkOnly bool
    fs := flag.NewFlagSet("update", flag.ContinueOnError)
    fs.BoolVar(&checkOnly, "check-only", false, "Only check for updates, don't install")
    fs.SetOutput(os.Stderr)
    _ = fs.Parse(args)

    release, err := updater.CheckLatestRelease()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to check for updates: %v\n", err)
        os.Exit(1)
    }

    if release.Version == currentVersion {
        fmt.Printf("You are already running the latest version: %s\n", currentVersion)
        return
    }

    fmt.Printf("New version available: %s (current: %s)\n", release.Version, currentVersion)

    if checkOnly {
        fmt.Println("Update check complete. Run 'sortpath update' to install the new version.")
        return
    }

    if !updater.IsInstalled() {
        fmt.Fprintf(os.Stderr, "Error: sortpath was not installed via the install command.\n")
        fmt.Fprintf(os.Stderr, "Please reinstall manually or run 'sortpath install' first.\n")
        os.Exit(1)
    }

    fmt.Printf("Downloading and installing version %s...\n", release.Version)
    if err := updater.UpdateBinary(release); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to install update: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Successfully updated to version %s!\n", release.Version)
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

func userHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return h
}

func userBinFallbackDir() string {
	h := userHomeDir()
	candidates := []string{
		filepath.Join(h, "bin"),
		filepath.Join(h, ".local", "bin"),
	}
	for _, d := range candidates {
		// Return first candidate; we'll create if needed
		return d
	}
	return ""
}

func pathContainsDir(dir string) bool {
	pathEnv := os.Getenv("PATH")
	for _, p := range strings.Split(pathEnv, ":") {
		if p == dir {
			return true
		}
	}
	return false
}

func addDirToShellPATH(dir string) (profilePath string, added bool, err error) {
	shell := filepath.Base(os.Getenv("SHELL"))
	h := userHomeDir()
	snippet := fmt.Sprintf("\n# Added by sortpath on %s\nexport PATH=\"%s:$PATH\"\n", time.Now().Format(time.RFC3339), dir)
	switch shell {
	case "zsh":
		profilePath = filepath.Join(h, ".zshrc")
	case "bash":
		// Prefer bash_profile on macOS
		pf := filepath.Join(h, ".bash_profile")
		if _, statErr := os.Stat(pf); statErr == nil {
			profilePath = pf
		} else {
			profilePath = filepath.Join(h, ".bashrc")
		}
	default:
		// Fallback to .profile
		profilePath = filepath.Join(h, ".profile")
	}
	// Read existing if exists and check if already contains dir
	if b, readErr := os.ReadFile(profilePath); readErr == nil {
		if strings.Contains(string(b), dir) {
			return profilePath, false, nil
		}
	}
	f, openErr := os.OpenFile(profilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if openErr != nil {
		return profilePath, false, openErr
	}
	defer f.Close()
	if _, werr := f.WriteString(snippet); werr != nil {
		return profilePath, false, werr
	}
	return profilePath, true, nil
}
