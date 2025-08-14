package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

type CLIOptions struct {
    APIKey  string
    APIBase string
    Model   string
    Tree    string
}

type Config struct {
    APIKey  string `yaml:"api_key"`
    APIBase string `yaml:"api_base"`
    Model   string `yaml:"model"`
    Tree    string `yaml:"tree"`
    // Whether to suppress the interactive install prompt on startup
    InstallPromptDisabled bool   `yaml:"install_prompt_disabled"`
    // Where the binary was installed (if installed)
    InstalledPath         string `yaml:"installed_path"`
    // Skip automatic update checks
    DisableAutoUpdate     bool   `yaml:"disable_auto_update"`
}

var configPath = filepath.Join(os.Getenv("HOME"), ".config", "sortpath", "config.yaml")

func Load() (*Config, error) {
    f, err := os.Open(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            return &Config{}, nil
        }
        return nil, err
    }
    defer f.Close()
    var c Config
    dec := yaml.NewDecoder(f)
    if err := dec.Decode(&c); err != nil {
        return nil, err
    }
    return &c, nil
}

func Save(c *Config) error {
    dir := filepath.Dir(configPath)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }
    f, err := os.Create(configPath)
    if err != nil {
        return err
    }
    defer f.Close()
    enc := yaml.NewEncoder(f)
    return enc.Encode(c)
}

func Set(key, value string) error {
    c, _ := Load()
    switch key {
    case "api-key":
        c.APIKey = value
    case "api_base", "api-base":
        c.APIBase = value
    case "model":
        c.Model = value
    case "tree":
        c.Tree = value
    case "install-prompt-disabled":
        if value == "true" || value == "1" || value == "yes" || value == "y" {
            c.InstallPromptDisabled = true
        } else {
            c.InstallPromptDisabled = false
        }
    case "installed-path":
        c.InstalledPath = value
    case "disable-auto-update":
        if value == "true" || value == "1" || value == "yes" || value == "y" {
            c.DisableAutoUpdate = true
        } else {
            c.DisableAutoUpdate = false
        }
    default:
        return errors.New("unknown config key")
    }
    return Save(c)
}

func Get(key string) (string, error) {
    c, _ := Load()
    switch key {
    case "api-key":
        return c.APIKey, nil
    case "api_base", "api-base":
        return c.APIBase, nil
    case "model":
        return c.Model, nil
    case "tree":
        return c.Tree, nil
    case "install-prompt-disabled":
        if c.InstallPromptDisabled {
            return "true", nil
        }
        return "false", nil
    case "installed-path":
        return c.InstalledPath, nil
    case "disable-auto-update":
        if c.DisableAutoUpdate {
            return "true", nil
        }
        return "false", nil
    default:
        return "", errors.New("unknown config key")
    }
}

func Remove(key string) error {
    c, _ := Load()
    switch key {
    case "api-key":
        c.APIKey = ""
    case "api_base", "api-base":
        c.APIBase = ""
    case "model":
        c.Model = ""
    case "tree":
        c.Tree = ""
    case "install-prompt-disabled":
        c.InstallPromptDisabled = false
    case "installed-path":
        c.InstalledPath = ""
    case "disable-auto-update":
        c.DisableAutoUpdate = false
    default:
        return errors.New("unknown config key")
    }
    return Save(c)
}

func (c *Config) ToMap() map[string]string {
    return map[string]string{
        "api-key":  c.APIKey,
        "api-base": c.APIBase,
        "model":    c.Model,
        "tree":     c.Tree,
        "install-prompt-disabled": func() string { if c.InstallPromptDisabled { return "true" } else { return "false" } }(),
        "installed-path":          c.InstalledPath,
        "disable-auto-update":     func() string { if c.DisableAutoUpdate { return "true" } else { return "false" } }(),
    }
}

// CLI flag > ENV > config.yaml
func ResolveConfig(opts CLIOptions) (*Config, error) {
    c, _ := Load()
    // ENV fallback
    if opts.APIKey == "" {
        opts.APIKey = os.Getenv("OPENAI_API_KEY")
    }
    if opts.APIBase == "" {
        opts.APIBase = os.Getenv("OPENAI_API_BASE")
    }
    if opts.Model == "" {
        opts.Model = os.Getenv("OPENAI_MODEL")
    }
    if opts.Tree == "" {
        opts.Tree = os.Getenv("SORTPATH_FOLDER_TREE")
    }
    // Config fallback
    if opts.APIKey == "" {
        opts.APIKey = c.APIKey
    }
    if opts.APIBase == "" {
        opts.APIBase = c.APIBase
    }
    if opts.Model == "" {
        opts.Model = c.Model
    }
    if opts.Tree == "" {
        opts.Tree = c.Tree
    }
    if opts.APIKey == "" || opts.APIBase == "" || opts.Model == "" || opts.Tree == "" {
        return nil, fmt.Errorf("missing required config (api-key, api-base, model, tree)")
    }
    return &Config{
        APIKey:  opts.APIKey,
        APIBase: opts.APIBase,
        Model:   opts.Model,
        Tree:    opts.Tree,
    }, nil
}