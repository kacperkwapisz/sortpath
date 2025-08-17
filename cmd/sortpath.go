package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kacperkwapisz/sortpath/internal/ai"
	"github.com/kacperkwapisz/sortpath/internal/config"
	"github.com/kacperkwapisz/sortpath/internal/fs"
	"github.com/kacperkwapisz/sortpath/internal/updater"
	"github.com/kacperkwapisz/sortpath/pkg/api"
	"github.com/kacperkwapisz/sortpath/pkg/cli"
)

var Version = "dev"

func main() {
    args := os.Args[1:]
    if len(args) == 0 || (len(args) == 1 && (args[0] == "-h" || args[0] == "--help")) {
        cli.PrintHelp(Version)
        return
    }

    // Version flag
    if len(args) == 1 && (args[0] == "-v" || args[0] == "--version") {
        fmt.Printf("🔍 sortpath version %s\n", Version)
        return
    }

    // Install subcommand
    if args[0] == "install" {
        cli.HandleInstallCommand(args[1:])
        return
    }

    // Config subcommand
    if args[0] == "config" {
        cli.HandleConfigCommand(args[1:])
        return
    }

    // Update subcommand
    if args[0] == "update" {
        cli.HandleUpdateCommand(args[1:], Version)
        return
    }

    // If the first argument is not "config" and not a quoted description, print help
    if len(args) == 1 && (args[0] == "list" || args[0] == "set" || args[0] == "get" || args[0] == "remove") {
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
        cli.PrintHelp(Version)
        os.Exit(1)
    }

    // First-run install prompt (non-blocking in non-interactive environments)
    maybePromptInstall()

    // Check for updates (non-blocking)
    if Version != "dev" {
        go checkForUpdates()
    }

    // Parse CLI flags and positional
    opts, desc := cli.ParseArgs(args)
    if desc == "" {
        fmt.Fprintf(os.Stderr, "Missing file description.\n")
        cli.PrintHelp(Version)
        os.Exit(1)
    }
    conf, err := config.ResolveConfig(opts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Config error: %v\n", err)
        os.Exit(1)
    }

    tree, err := fs.Tree(conf.TreePath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Folder tree error: %v\n", err)
        os.Exit(1)
    }

    prompt := ai.BuildPrompt(tree, desc)
    resp, err := api.QueryLLM(conf, prompt)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ API error: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(resp.Path)
    fmt.Printf("Reason: %s\n", resp.Reason)
}

func checkForUpdates() {
    if Version == "dev" {
        return
    }

    // Auto-update checks are now always enabled (following YAGNI principle)

    // Check if it's been at least 1 minute since last check
    lastCheck, err := updater.GetLastUpdateCheck()
    if err != nil {
        // On error, proceed as if never checked
        lastCheck = time.Time{}
    }
    
    now := time.Now()
    if !lastCheck.IsZero() && now.Sub(lastCheck) < 1*time.Minute {
        return // Already checked within last minute
    }

    release, err := updater.CheckLatestRelease()
    if err != nil {
        // Silently fail, but update last check time to prevent rapid retries
        _ = updater.SetLastUpdateCheck(now)
        return
    }

    // Update the last check time
    _ = updater.SetLastUpdateCheck(now)

    if release.Version != Version {
        header, instruction := updater.FormatUpdateNotification(release.Version, Version, true)
        fmt.Fprintf(os.Stderr, "\n%s\n", header)
        fmt.Fprintf(os.Stderr, "%s\n\n", instruction)
    }
}

// Add version info to help output
func init() {
}

func maybePromptInstall() {
    // Always show install prompt if not already installed (following YAGNI principle)

    // If executable is already in PATH dir, skip
    execPath, err := os.Executable()
    if err != nil {
        return
    }
    execDir := filepath.Dir(execPath)
    if cliIsDirInPATH(execDir) {
        return
    }

    // If stdin is not a terminal, skip prompt
    if fi, _ := os.Stdin.Stat(); (fi.Mode() & os.ModeCharDevice) == 0 {
        return
    }

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("📦 Install sortpath to /usr/local/bin so you can run it from anywhere? [Y/n]: ")
    answer, _ := reader.ReadString('\n')
    answer = strings.TrimSpace(strings.ToLower(answer))
    if answer == "" || answer == "y" || answer == "yes" {
        // Attempt install
        os.Args = append([]string{os.Args[0], "install"}, os.Args[1:]...)
        cli.HandleInstallCommand([]string{})
        return
    }
    // User declined installation - no need to track this anymore
}

func cliIsDirInPATH(dir string) bool {
    // mirror of pathContainsDir in cli package, but unexported there; simple recheck here
    pathEnv := os.Getenv("PATH")
    for _, p := range strings.Split(pathEnv, ":") {
        if p == dir {
            return true
        }
    }
    return false
}