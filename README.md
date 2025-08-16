# 🗂️ sortpath — AI-Powered Folder Recommendation CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/kacperkwapisz/sortpath)](https://goreportcard.com/report/github.com/kacperkwapisz/sortpath)
[![Release](https://img.shields.io/github/release/kacperkwapisz/sortpath.svg)](https://github.com/kacperkwapisz/sortpath/releases)
[![Downloads](https://img.shields.io/github/downloads/kacperkwapisz/sortpath/total.svg)](https://github.com/kacperkwapisz/sortpath/releases)
[![Contributors](https://img.shields.io/github/contributors/kacperkwapisz/sortpath.svg)](https://github.com/kacperkwapisz/sortpath/graphs/contributors)
[![Issues](https://img.shields.io/github/issues/kacperkwapisz/sortpath.svg)](https://github.com/kacperkwapisz/sortpath/issues)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Never wonder where to put your files again.** Let AI analyze your folder structure and recommend the perfect location.

sortpath reads your existing folder hierarchy and uses AI to suggest where files should be stored based on their description—keeping your digital workspace organized and logical.

<!-- ![Demo](assets/demo.gif) -->
<!-- *Watch sortpath analyze folder structure and recommend the perfect location in seconds* -->

🏷️ **Keywords:** `file organization` `folder structure` `AI assistant` `CLI tool` `directory management` `file sorting` `digital organization` `workspace automation` `OpenAI` `GPT` `productivity tool` `cross-platform`

---

## 🚀 Quick Start

Get intelligent folder recommendations in 3 steps:

```bash
# 1. Install sortpath
curl -fsSL https://raw.githubusercontent.com/kacperkwapisz/sortpath/main/install.sh | bash

# 2. Configure your API credentials
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_API_BASE="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-3.5-turbo"

# 3. Get folder recommendations
sortpath "Invoice PDFs from TechCorp project, May 2024"
# Output:
# /01_PROJECTS/2024/TechCorp/Invoices
# Reason: Project-specific invoices are stored in year-based subfolders under Projects.
```

---

## ✨ Features

- 🧠 **AI-powered analysis** — Uses any OpenAI-compatible API (OpenAI, Anthropic, local models)
- 📁 **Reads your structure** — Analyzes your existing folder hierarchy automatically
- ⚡ **Lightning fast** — Single binary with zero dependencies
- 🔧 **Flexible config** — CLI flags, environment variables, or config files
- 🔄 **Self-updating** — Built-in update and install commands
- 🔒 **Privacy-focused** — Only folder structure and file descriptions are sent to AI

---

## 📦 Installation

### Automated Install (Recommended)

**All platforms (Linux, macOS, Windows):**

```bash
curl -fsSL https://raw.githubusercontent.com/kacperkwapisz/sortpath/main/install.sh | bash
```

The install script automatically:

- Detects your OS and architecture
- Downloads the latest release from GitHub
- Installs to `/usr/local/bin` (or `~/.local/bin` as fallback)
- Works on Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64)

### Other Installation Methods

<details>
<summary>Manual Download</summary>

Download pre-built binaries from [GitHub Releases](https://github.com/kacperkwapisz/sortpath/releases):

- `sortpath-linux-amd64`
- `sortpath-linux-arm64`
- `sortpath-darwin-amd64`
- `sortpath-darwin-arm64`
- `sortpath-windows-amd64.exe`

</details>

<details>
<summary>Self-Install (if you already have sortpath)</summary>

```bash
sortpath install
```

</details>

<details>
<summary>From Source</summary>

```bash
git clone https://github.com/kacperkwapisz/sortpath.git
cd sortpath
go build -o sortpath ./cmd/sortpath.go
```

Requires Go 1.21+

</details>

<details>
<summary>Go Install</summary>

```bash
go install github.com/kacperkwapisz/sortpath@latest
```

</details>

---

## 🎯 Usage

### Basic Usage

sortpath analyzes your folder structure and recommends where files should go:

```bash
# Basic recommendation
sortpath "Meeting notes from Q4 planning session"
# Output:
# /03_ADMIN/MEETINGS/2024/Q4
# Reason: Meeting notes are organized by year and quarter for easy retrieval.

# Project-specific files
sortpath "Logo mockups for BrandX redesign project"
# Output:
# /01_PROJECTS/2024/BrandX/Design/Logos
# Reason: Project-specific design assets are grouped under the project folder.
```

### How It Works

1. **Reads folder structure** — sortpath scans your current directory (or specified path) to understand your organizational system
2. **Analyzes with AI** — Sends the folder structure + your file description to an AI model
3. **Returns recommendation** — Gets back a specific folder path and explanation

### Complete Setup Example

```bash
# Set up configuration
export OPENAI_API_KEY="sk-your-key-here"
export OPENAI_API_BASE="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-3.5-turbo"

# Point to your folder structure (optional - defaults to current directory)
export SORTPATH_FOLDER_TREE="/path/to/your/project/structure"

# Get recommendations
sortpath "Raw video footage from product demo shoot"
# Output based on YOUR folder structure

sortpath "Python automation scripts for data processing"
# Output based on YOUR folder structure
```

### CLI Options

| Flag         | Description               | Example                                |
| ------------ | ------------------------- | -------------------------------------- |
| `--api-key`  | OpenAI-compatible API key | `--api-key sk-xxx`                     |
| `--api-base` | API base URL              | `--api-base https://api.openai.com/v1` |
| `--model`    | Model name                | `--model gpt-4`                        |
| `--tree`     | Path to folder structure  | `--tree ~/Documents/structure`         |

### Subcommands

| Command   | Description                                |
| --------- | ------------------------------------------ |
| `install` | Install binary to PATH directory           |
| `update`  | Update to latest version from GitHub       |
| `config`  | Manage configuration (set/get/remove/list) |

---

## ⚙️ Configuration

Configure sortpath using any combination of:

### 1. CLI Flags

```bash
sortpath --api-key sk-xxx --model gpt-4 --tree ~/my-folders "File description"
```

### 2. Environment Variables

```bash
export OPENAI_API_KEY="sk-xxx"
export OPENAI_API_BASE="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-3.5-turbo"
export SORTPATH_FOLDER_TREE="~/Documents/structure"
```

### 3. Config File (`~/.config/sortpath/config.yaml`)

```bash
# Set values
sortpath config set api-key sk-xxx
sortpath config set api-base https://api.openai.com/v1
sortpath config set model gpt-3.5-turbo
sortpath config set tree ~/Documents/structure

# View current config
sortpath config list

# Get specific value
sortpath config get api-key
```

**Priority order:** CLI flags → Environment variables → Config file

### Required Configuration

sortpath needs these three values to work:

- `api-key` — Your OpenAI-compatible API key
- `api-base` — API endpoint URL
- `model` — Model name to use

Optional configuration:

- `tree` — Path to folder structure (defaults to current directory)

---

## 💡 Examples

### Real-world Use Cases

The AI understands context and suggests appropriate locations:

```bash
# Creative work
sortpath "Logo concepts for TechCorp rebrand project"
# → /CLIENTS/TechCorp/Design/Logos/concepts

# Development files
sortpath "Unit tests for authentication module"
# → /src/auth/tests

# Personal organization
sortpath "Photos from Berlin vacation, Summer 2024"
# → /Photos/2024/Travel/Berlin

# Business documents
sortpath "Q3 financial reports and budget spreadsheets"
# → /Finance/2024/Q3/Reports
```

### Compatible with Any Structure

sortpath works with your existing folder organization:

<details>
<summary>📁 PARA Method</summary>

```
/Projects
/Areas
/Resources
/Archive
```

</details>

<details>
<summary>📁 Creative Professional</summary>

```
/01_PROJECTS
  /2024
  /2023
/02_CLIENTS
/03_RESOURCES
/04_DESIGN
  /LOGOS
  /MOCKUPS
  /TEMPLATES
/05_MEDIA
/06_ARCHIVE
```

</details>

<details>
<summary>📁 Software Development</summary>

```
/src
/tests
/docs
/scripts
/config
/assets
/build
```

</details>

### Alternative AI Providers

sortpath works with any OpenAI-compatible API:

```bash
# Anthropic Claude
export OPENAI_API_BASE="https://api.anthropic.com/v1"
export OPENAI_MODEL="claude-3-sonnet-20240229"

# Local models (ollama, etc.)
export OPENAI_API_BASE="http://localhost:11434/v1"
export OPENAI_MODEL="llama2"

# Other providers
export OPENAI_API_BASE="https://api.groq.com/openai/v1"
export OPENAI_MODEL="mixtral-8x7b-32768"
```

---

## 🛠️ Troubleshooting

### Common Issues

**"❌ Config error: missing required config"**

```bash
# Make sure all three required values are set:
sortpath config list
# Should show: api-key, api-base, model (tree is optional)
```

**"❌ API error: 401 Unauthorized"**

```bash
# Check your API key is valid
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
```

**"❌ Folder tree error: no such file or directory"**

```bash
# Set tree to an existing directory path
sortpath config set tree ~/Documents
# OR remove tree config to use current directory
sortpath config remove tree
```

**"❌ API error: rate limit exceeded"**

- Wait a few minutes before trying again
- Consider upgrading your API plan
- Use a different model with higher limits

**"No response from model"**

- Check if the model name is correct for your provider
- Verify your API base URL is correct
- Try with a different model (e.g., `gpt-3.5-turbo`)

**Binary not found after install**

```bash
# Check if install directory is in PATH
echo $PATH | grep -o "/usr/local/bin\|$HOME/.local/bin\|$HOME/bin"

# If not found, add to shell profile
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Permission denied on install**

```bash
# Install script will automatically try fallback locations
# OR manually specify install directory
INSTALL_DIR="$HOME/bin" curl -fsSL https://raw.githubusercontent.com/kacperkwapisz/sortpath/main/install.sh | bash
```

### Getting Help

- 📖 **Documentation issues?** [Open a docs issue](https://github.com/kacperkwapisz/sortpath/issues/new?labels=documentation)
- 🐛 **Found a bug?** [Report it here](https://github.com/kacperkwapisz/sortpath/issues/new?labels=bug)
- 💬 **Questions?** [Start a discussion](https://github.com/kacperkwapisz/sortpath/discussions)
- 📧 **Private issues?** Email: kacper@kacperkwapisz.com

---

## 🔄 Updating

Stay up-to-date with the latest features:

```bash
# Check for updates
sortpath update --check-only

# Update to latest version
sortpath update
```

---

## 🤝 Contributing

We love contributions! Here's how to get involved:

- 🐛 **Found a bug?** [Open an issue](https://github.com/kacperkwapisz/sortpath/issues/new)
- 💡 **Have an idea?** [Start a discussion](https://github.com/kacperkwapisz/sortpath/discussions)
- 🛠️ **Want to contribute code?** Check out [`CONTRIBUTING.md`](CONTRIBUTING.md)

### Community Stats

- 📊 **Active Issues:** [![Issues](https://img.shields.io/github/issues/kacperkwapisz/sortpath.svg)](https://github.com/kacperkwapisz/sortpath/issues)
- 👥 **Contributors:** [![Contributors](https://img.shields.io/github/contributors/kacperkwapisz/sortpath.svg)](https://github.com/kacperkwapisz/sortpath/graphs/contributors)
- 📈 **Total Downloads:** [![Downloads](https://img.shields.io/github/downloads/kacperkwapisz/sortpath/total.svg)](https://github.com/kacperkwapisz/sortpath/releases)
- ⭐ **GitHub Stars:** [![Stars](https://img.shields.io/github/stars/kacperkwapisz/sortpath.svg?style=social)](https://github.com/kacperkwapisz/sortpath)

### Recent Activity

![GitHub commit activity](https://img.shields.io/github/commit-activity/m/kacperkwapisz/sortpath)
![GitHub last commit](https://img.shields.io/github/last-commit/kacperkwapisz/sortpath)

---

## 📄 License

Distributed under the MIT License. See [`LICENSE`](LICENSE) for details.

---

## 👨‍💻 About

Built with ❤️ by [@kacperkwapisz](https://github.com/kacperkwapisz)

**Stack:** Go • OpenAI-compatible APIs • Cross-platform CLI

---

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=kacperkwapisz/sortpath&type=Date)](https://star-history.com/#kacperkwapisz/sortpath&Date)
