package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kacperkwapisz/sortpath/internal/config"
	"github.com/kacperkwapisz/sortpath/internal/updater"
)

// CLIOptions is now defined in the config package

func ParseArgs(args []string) (config.CLIOptions, string) {
    var opts config.CLIOptions
    fs := flag.NewFlagSet("sortpath", flag.ContinueOnError)
    fs.StringVar(&opts.APIKey, "api-key", "", "OpenAI-compatible API key")
    fs.StringVar(&opts.APIBase, "api-base", "", "API base URL")
    fs.StringVar(&opts.Model, "model", "", "Model name")
    fs.StringVar(&opts.TreePath, "tree", "", "Path to folder tree file")
    fs.StringVar(&opts.LogLevel, "log-level", "", "Log level (debug, info, error)")
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
  --api-key    OpenAI-compatible API key
  --api-base   API base URL (e.g. https://api.openai.com/v1)
  --model      Model name (e.g. gpt-3.5-turbo)
  --tree       Path to folder tree file
  --log-level  Log level (debug, info, error)
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
        PrintHelp("dev")
        return
    }
    switch args[0] {
    case "set":
        if len(args) != 3 {
            fmt.Println("Usage: sortpath config set <key> <value>")
            return
        }
        err := setConfigValue(args[1], args[2])
        if err != nil {
            fmt.Fprintf(os.Stderr, "❌ Config set error: %v\n", err)
            os.Exit(1)
        }
    case "get":
        if len(args) != 2 {
            fmt.Println("Usage: sortpath config get <key>")
            return
        }
        val, err := getConfigValue(args[1])
        if err != nil {
            fmt.Fprintf(os.Stderr, "❌ Config get error: %v\n", err)
            os.Exit(1)
        }
        fmt.Println(val)
    case "remove":
        if len(args) != 2 {
            fmt.Println("Usage: sortpath config remove <key>")
            return
        }
        err := removeConfigValue(args[1])
        if err != nil {
            fmt.Fprintf(os.Stderr, "❌ Config remove error: %v\n", err)
            os.Exit(1)
        }
    case "list":
        conf, err := config.Load()
        if err != nil {
            fmt.Fprintf(os.Stderr, "❌ Config list error: %v\n", err)
            os.Exit(1)
        }
        configMap := map[string]string{
            "api-key":   conf.APIKey,
            "api-base":  conf.APIBase,
            "model":     conf.Model,
            "tree-path": conf.TreePath,
            "log-level": conf.LogLevel,
        }
        for k, v := range configMap {
            fmt.Printf("%s: %s\n", k, v)
        }
    default:
        PrintHelp("dev")
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
        fmt.Fprintf(os.Stderr, "❌ Cannot determine current executable path: %v\n", err)
        os.Exit(1)
    }

    destPath := filepath.Join(destDir, "sortpath")
    if !force {
        if _, err := os.Stat(destPath); err == nil {
            fmt.Fprintf(os.Stderr, "⚠️ Destination already has sortpath: %s (use --force to overwrite)\n", destPath)
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

            // Ensure PATH contains fallbackDir; if not, attempt to add to shell profile
            if !pathContainsDir(fallbackDir) {
                profilePath, added, addErr := addDirToShellPATH(fallbackDir)
                if addErr == nil && added {
                    fmt.Printf("Installed sortpath to %s and added it to PATH in %s. Restart your shell or run: source %s\n", userDest, profilePath, profilePath)
                } else {
                    fmt.Printf("Installed sortpath to %s. Add it to your PATH by adding this to your shell profile:\n\n    export PATH=\"%s:$PATH\"\n\nThen restart your terminal.\n", userDest, fallbackDir)
                }
            } else {
                fmt.Printf("✅ Installed sortpath to %s\n", userDest)
            }
            return
        }
        fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
        fmt.Fprintf(os.Stderr, "Try: sudo cp %q %q\n", srcPath, destPath)
        os.Exit(1)
    }
    // Make executable
    _ = os.Chmod(destPath, 0755)

    // Installation complete
    fmt.Printf("✅ Installed sortpath to %s\n", destPath)
}

func HandleUpdateCommand(args []string, currentVersion string) {
    var checkOnly bool
    fs := flag.NewFlagSet("update", flag.ContinueOnError)
    fs.BoolVar(&checkOnly, "check-only", false, "Only check for updates, don't install")
    fs.SetOutput(os.Stderr)
    _ = fs.Parse(args)

    release, err := updater.CheckLatestRelease()
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to check for updates: %v\n", err)
        os.Exit(1)
    }

    if release.Version == currentVersion {
        fmt.Printf("✅ You are already running the latest version: %s\n", currentVersion)
        return
    }

    header, instruction := updater.FormatUpdateNotification(release.Version, currentVersion, false)
    fmt.Println(header)

    if checkOnly {
        fmt.Println(instruction)
        return
    }

    if !updater.IsInstalled() {
        fmt.Fprintf(os.Stderr, "❌ Error: sortpath was not installed via the install command.\n")
        fmt.Fprintf(os.Stderr, "Please reinstall manually or run 'sortpath install' first.\n")
        os.Exit(1)
    }

    fmt.Printf("📦 Downloading and installing version %s...\n", release.Version)
    if err := updater.UpdateBinary(release); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to install update: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("✅ Successfully updated to version %s!\n", release.Version)
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

func setConfigValue(key, value string) error {
    c, _ := config.Load()
    switch key {
    case "api-key":
        c.APIKey = value
    case "api-base":
        c.APIBase = value
    case "model":
        c.Model = value
    case "tree-path":
        c.TreePath = value
    case "log-level":
        c.LogLevel = value
    default:
        return fmt.Errorf("unknown config key: %s", key)
    }
    return config.Save(c)
}

func getConfigValue(key string) (string, error) {
    c, _ := config.Load()
    switch key {
    case "api-key":
        return c.APIKey, nil
    case "api-base":
        return c.APIBase, nil
    case "model":
        return c.Model, nil
    case "tree-path":
        return c.TreePath, nil
    case "log-level":
        return c.LogLevel, nil
    default:
        return "", fmt.Errorf("unknown config key: %s", key)
    }
}

func removeConfigValue(key string) error {
    c, _ := config.Load()
    switch key {
    case "api-key":
        c.APIKey = ""
    case "api-base":
        c.APIBase = ""
    case "model":
        c.Model = ""
    case "tree-path":
        c.TreePath = ""
    case "log-level":
        c.LogLevel = ""
    default:
        return fmt.Errorf("unknown config key: %s", key)
    }
    return config.Save(c)
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