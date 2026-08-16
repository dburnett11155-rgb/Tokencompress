# ⚡ tokencompress (v1.1.0)

> A zero-dependency, sub-millisecond Go CLI and MCP sidecar that prunes raw tool outputs (JSON, terminal logs, HTML) before they enter your AI agent's context window.

Cut LLM context token consumption by **60% to 80%** without using a second LLM summarization turn.

[![Buy TokenCompress Pro](https://img.shields.io/badge/Gumroad-TokenCompress%20Pro%20%2429-2EA44F?style=for-the-badge&logo=gumroad)](https://burnettwave53.gumroad.com/l/nixgh)

---

## ⚡ Get TokenCompress Pro (9)

For instant pre-compiled binaries and turnkey configuration files, grab the **Pro Kit on Gumroad**:
👉 **[Download TokenCompress Pro (9)](https://burnettwave53.gumroad.com/l/nixgh)**

* **Ready-to-Use Executables:** Windows (), macOS (Apple Silicon/Intel), Linux
* **Pre-Configured MCP Settings:** Drop-in configs for Claude Desktop, Cursor, and AutoGen
* **Commercial License:** Use across unlimited team or enterprise workflows

---

## 💡 The Problem

When AI agents execute tools (calling APIs, running terminal commands, or scraping web pages), they receive thousands of lines of raw, unparsed data:
* **Massive JSON Payloads:** A 1,000-item array floods the prompt with 15,000+ tokens.
* **Verbose Stack Traces:** Framework noise drowning out root exceptions.
* **Raw HTML:** CSS, scripts, and navigation menus bloating context windows.

---

## ⚡ The Solution

 acts as a high-speed filter between your tools and your model:

* **JSON Truncation:** Keeps representative schema examples and replaces redundant array items with metadata (: N).
* **Log Pruning:** Strips internal framework paths and isolates root exceptions and user code.
* **HTML Cleaning:** Converts raw web scrapes to clean Markdown/text.
* **Duplicate Loop Detection:** Hashes tool outputs per session and warns models if they receive identical tool results repeatedly.

---

## 🚀 Quickstart (Open Source CLI)

### 1. Installation
make build
sudo mv tokencompress /usr/local/bin/

### 2. CLI Pipe Mode
# Compress a large JSON response
cat large_response.json | tokencompress --mode json

# Prune a verbose stack trace log
cat app.log | tokencompress --mode log --log-internal-marker "mycompany/internal"

# Clean raw HTML
curl -s https://example.com | tokencompress --mode html

---

## 📊 Benchmarks

| Input Payload | Raw Token Count | Compressed Token Count | Token Reduction | Execution Time |
| :--- | :--- | :--- | :--- | :--- |
| **JSON Array (500 items)** | ~12,500 tokens | ~850 tokens | **93.2%** |  |
| **Python Stack Trace** | ~3,200 tokens | ~410 tokens | **87.1%** |  |
| **HTML Web Scrape** | ~18,000 tokens | ~2,100 tokens | **88.3%** |  |

---

## 💼 Custom Integrations & Consulting

Need a custom high-performance MCP proxy, specialized parsers for enterprise tools, or team context setup?

* **Custom Setup Gigs:** 50 - 00 per custom integration setup.
* **Contact:** Open an issue on this repository or reach out via GitHub profile.

---

## 📜 License
MIT License.
