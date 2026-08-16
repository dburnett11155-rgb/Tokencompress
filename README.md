# ⚡ tokencompress (v1.1.0)

> A zero-dependency, sub-millisecond Go CLI and MCP sidecar that prunes raw tool outputs (JSON, terminal logs, HTML) before they enter your AI agent's context window.

Cut LLM context token consumption by **60% to 80%** without using a second LLM summarization turn.

[![Buy TokenCompress Pro](https://img.shields.io/badge/Gumroad-TokenCompress%20Pro%20%2429-2EA44F?style=for-the-badge&logo=gumroad)](https://burnettwave53.gumroad.com/l/nixgh)

---

## ⚡ Get TokenCompress Pro ($29)

For instant pre-compiled binaries and turnkey configuration files, grab the **Pro Kit on Gumroad**:
👉 **[Download TokenCompress Pro ($29)](https://burnettwave53.gumroad.com/l/nixgh)**

* **Ready-to-Use Executables:** Windows (`.exe`), macOS (Apple Silicon & Intel), and Linux (`x86_64` / `arm64`).
* **Pre-Configured MCP Settings:** Drop-in `claude_desktop_config.json` templates for Claude Desktop, Cursor, and Windsurf.
* **Turnkey Setup Scripts:** 1-click `install.sh` and `install.ps1` scripts for instant global setup without Go toolchain dependencies.
* **Commercial License:** Unlimited commercial use across individual, team, or enterprise workflows.

---

## 💡 The Problem

When AI agents execute tools (calling APIs, running terminal commands, or scraping web pages), they receive thousands of lines of raw, unparsed data:
* **Massive JSON Payloads:** A 1,000-item array floods the prompt with 15,000+ unnecessary tokens.
* **Verbose Stack Traces:** Hundreds of lines of framework boilerplate drowning out root exceptions.
* **Raw HTML:** Inline CSS, scripts, SVGs, and navigation menus bloating context windows.

---

## ⚡ The Solution

`tokencompress` acts as a high-speed filter between your tools and your model:

* **JSON Truncation (`tokencompress_json`):** Retains representative schema examples and replaces redundant array items with metadata (`_omitted_items_count: N`).
* **Log Pruning (`tokencompress_log`):** Strips internal framework paths and isolates user stack traces and root exceptions.
* **HTML Cleaning (`tokencompress_html`):** Strips scripts, styles, and navigation tags to output clean, structured Markdown text.
* **Duplicate Loop Detection:** Hashes tool outputs per session and warns models if they receive identical tool results repeatedly.

---

## 🚀 Quickstart

### Option A: Install Pre-Built Binaries (Pro)
Download the `$29 Pro Kit` from [Gumroad](https://burnettwave53.gumroad.com/l/nixgh), extract the package, and run the included installer:
```bash
# macOS / Linux
chmod +x install.sh && ./install.sh

# Windows (PowerShell)
.\install.ps1
```

### Option B: Build From Source (Open Source CLI)
Requires Go 1.22 or higher installed:

```bash
git clone [https://github.com/dburnett11155-rgb/Tokencompress.git](https://github.com/dburnett11155-rgb/Tokencompress.git)
cd Tokencompress
make build
sudo mv tokencompress /usr/local/bin/
```

---

## 💻 CLI Usage Examples

```bash
# Compress a large JSON response
cat large_response.json | tokencompress --mode json

# Prune a verbose stack trace log
cat app.log | tokencompress --mode log --log-internal-marker "mycompany/internal"

# Clean raw HTML scrapes
curl -s [https://example.com](https://example.com) | tokencompress --mode html
```

---

## 📊 Benchmarks

| Input Payload | Raw Token Count | Compressed Token Count | Token Reduction | Execution Time |
| :--- | :--- | :--- | :--- | :--- |
| **JSON Array (500 items)** | ~12,500 tokens | ~850 tokens | **93.2%** | < 0.4ms |
| **Python Stack Trace** | ~3,200 tokens | ~410 tokens | **87.1%** | < 0.2ms |
| **HTML Web Scrape** | ~18,000 tokens | ~2,100 tokens | **88.3%** | < 0.8ms |

---

## 💼 Custom Integrations & Consulting

Need a custom high-performance MCP proxy, specialized parsers for enterprise tools, or custom team context rules?

* **Custom Integration Services:** Custom agent & MCP workflow setup available ($250 – $1,000).
* **Contact:** Open an issue on this repository or reach out via GitHub profile.

---

## 📜 License
MIT License.
