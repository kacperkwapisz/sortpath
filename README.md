# sortpath

**sortpath** is a blazing-fast, AI-powered CLI tool that recommends perfect folder paths for your files based on your personal archive structure.  
Works with any OpenAI-compatible API, including OpenAI, OpenRouter, Azure OpenAI, and self-hosted local models.

## Features

- **AI-Powered Recommendations:** Uses the latest LLMs to suggest optimal storage paths for any file asset using your real folder tree.
- **Custom Archive Structure:** Easily load and update your personal folder structure from config or the filesystem.
- **Works Anywhere:** Use with OpenAI, Azure, OpenRouter, local server, or any OpenAI-compatible endpoint.
- **Configurable and Secure:** API keys and endpoint are managed via CLI flags, config file, `config` command, or environment variables.
- **Config Command:** Set, get, and list config options easily for effortless daily use (no need to repeat flags).
- **Blazing Fast, Single-Binary CLI:** Written in Go for maximum speed, tiny size, and zero dependencies.

## Example Usage

```
# First run will offer to install to /usr/local/bin so you can run `sortpath` from anywhere
sortpath "Berlin trip photos 2025"

# Manual install (later or if you declined the prompt)
sortpath install --path /usr/local/bin

# Use all saved settings
sortpath "Berlin trip photos 2025"
# Output:
# /03_PHOTOS/2025/Berlin_Trip
# Reason: Photos by year and event; keeps travel memories organized.

# Override default config for a single run
sortpath --api-base 'https://openrouter.ai/v1' --api-key 'sk-xxx' --model 'nous-hermes-2-mixtral' "PSD shoe mockups"

# Set persistent config values
sortpath config set api-key sk-xxxx
sortpath config set api-base https://openrouter.ai/v1
sortpath config set model gpt-3.5-turbo
sortpath config set tree ~/sorttree.txt

# Show current config
sortpath config list

# Remove a config value
sortpath config remove api-key
```

## Configuration

### Environment Variables

```
OPENAI_API_KEY=sk-yourkey
OPENAI_API_BASE=https://api.openai.com/v1
OPENAI_MODEL=gpt-4.1
SORTPATH_FOLDER_TREE=path/to/tree.txt
```

### CLI Flags

- `--api-key` : Override API key
- `--api-base` : Override API endpoint
- `--model` : Choose a model (e.g., gpt-4)
- `--tree` : Path to your latest folder tree file or config

### Config Command

Use the `config` command to set and manage persistent defaults interactively:

- Set: `sortpath config set key value`
- Get: `sortpath config get key`
- Remove: `sortpath config remove key`
- List all: `sortpath config list`

Stores config in `~/.config/sortpath/config.yaml`.

### Install

```
# Install to /usr/local/bin (may require sudo depending on permissions)
sortpath install

# Custom destination	sortpath install --path "$HOME/bin"
```

If you declined the first-run prompt and want it again:

```
sortpath config remove install-prompt-disabled
```

## Project Structure

```
sortpath/
├─ cmd/
├─ internal/
│   ├─ ai/
│   ├─ config/
│   └─ fs/
├─ pkg/
│   ├─ api/
│   └─ cli/
├─ assets/
├─ README.md
├─ go.mod
├─ go.sum
├─ .env.example
└─ .gitignore
```

## Getting Started

1. Clone repo and initialize:

   ```
   git clone https://github.com/yourusername/sortpath.git
   cd sortpath
   go mod tidy
   ```

2. Configure your `.env` file, use the `config` command, or set environment variables.

3. Run:
   ```
   go run ./cmd/sortpath.go "Describe your file here"
   ```

## Roadmap

- [x] Modular, OpenAI-compatible CLI
- [x] Interactive `config` command for persistent settings
- [ ] JSON/YAML config support for project and folder settings
- [ ] GUI or web frontend
- [ ] Team/workspace collaboration
- [ ] More LLM providers and offline support

## License

MIT

## Contributing

PRs and issues welcome! Please open an issue first to discuss major changes.
