# ⚡ tokencompress (v1.1.0)

> A zero-dependency, sub-millisecond Go CLI and MCP sidecar that prunes raw tool outputs (JSON, terminal logs, HTML) before they enter your AI agent's context window.

Cut LLM context token consumption by **60% to 80%** without using a second LLM summarization turn.

---

## 💡 The Problem

When AI agents execute tools (calling APIs, running terminal commands, or scraping web pages), they receive thousands of lines of raw, unparsed data:
* **Massive JSON Payloads:** A 1,000-item array floods the prompt with 15,000+ tokens.
* **Verbose Stack Traces:** Framework noise and node_modules paths drown out the root exception.
* **Raw HTML:** CSS, scripts, and navigation menus bloat the context window.

This leads to high API costs, slower response times, and **context rot**—where agents hallucinate or repeat tool calls mid-task.

---

## ⚡ The Solution

 acts as a high-speed, deterministic filter between your tools and your model:

* **JSON Truncation:** Keeps representative schema examples and replaces redundant array items with metadata (: N).
* **Log Pruning:** Strips internal framework paths and returns only the root error message, user file paths, and execution lines.
* **HTML Cleaning:** Strips scripts, styles, and navigation elements, converting content to readable text/markdown.
* **Duplicate Loop Detection:** Hashes tool outputs per session and prepends a warning header if an agent receives duplicate tool results twice in a row.

---

## 🚀 Quickstart

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

### 3. MCP (Model Context Protocol) Integration
Add  directly to your Claude Desktop config ():

{
  "mcpServers": {
    "tokencompress": {
      "command": "/usr/local/bin/tokencompress",
      "args": ["--mode", "mcp"]
    }
  }
}

---

## 📊 Benchmarks

| Input Payload | Raw Token Count | Compressed Token Count | Token Reduction | Execution Time |
| :--- | :--- | :--- | :--- | :--- |
| **JSON Array (500 items)** | ~12,500 tokens | ~850 tokens | **93.2%** |  |
| **Python Stack Trace** | ~3,200 tokens | ~410 tokens | **87.1%** |  |
| **HTML Web Scrape** | ~18,000 tokens | ~2,100 tokens | **88.3%** |  |

---

## 💼 Custom Integrations & Consulting

Need a custom high-performance MCP proxy, specialized parsers for enterprise tools, or custom token-optimization setups for your agent infrastructure?

* **Custom Setup Gigs:** 50 - 00 per custom integration setup.
* **Custom Log/AST Parsers:** 50 per custom domain parser.
* **Contact:** Open an issue on this repository or contact directly via GitHub profile.

---

## 📜 License
MIT License. Free to use in open-source and commercial agent setups.
