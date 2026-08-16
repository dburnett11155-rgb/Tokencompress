// tokencompress.go
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const version = "1.1.0"

// Default JSON truncation parameters (overridden by flags)
var (
	jsonMaxItems  = 5
	jsonKeepFirst = 3
)

// Custom internal path markers added via flags
var customInternalMarkers []string

// ---------------------------------------------------------------------------
// Session & Duplicate Interceptor (Feature D)
// ---------------------------------------------------------------------------

// Session tracks hashes of cleaned outputs per mode/tool.
// It is safe for concurrent use.
type Session struct {
	mu     sync.Mutex
	hashes map[string]string
}

// NewSession creates an empty session.
func NewSession() *Session {
	return &Session{hashes: make(map[string]string)}
}

// CheckAndStore computes the SHA-256 of cleaned output. If the hash matches
// the previous hash for the same key, it prepends a duplicate warning.
func (s *Session) CheckAndStore(key string, cleaned []byte) []byte {
	sum := sha256.Sum256(cleaned)
	h := hex.EncodeToString(sum[:])

	s.mu.Lock()
	defer s.mu.Unlock()

	if prev, ok := s.hashes[key]; ok && prev == h {
		warning := []byte("[TOKENCOMPRESS WARNING: TOOL RETURNED DUPLICATE OUTPUT. " +
			"DO NOT REPEAT THIS TOOL CALL WITH IDENTICAL PARAMETERS.]\n")
		return append(warning, cleaned...)
	}
	s.hashes[key] = h
	return cleaned
}

// LoadSessionFromFile loads a session cache from disk (optional CLI persistence).
func LoadSessionFromFile(path string) (*Session, error) {
	s := NewSession()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, &s.hashes); err != nil {
		return nil, err
	}
	return s, nil
}

// SaveSessionToFile saves the session cache to disk.
func (s *Session) SaveSessionToFile(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s.hashes, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ---------------------------------------------------------------------------
// Feature A: Deterministic JSON Array Truncator
// ---------------------------------------------------------------------------

// truncateJSON recursively truncates any array with >5 elements (or custom
// max items if flags are set). It keeps the first N elements (customizable)
// and appends a metadata object with the omitted count.
func truncateJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, child := range val {
			out[k] = truncateJSON(child)
		}
		return out

	case []interface{}:
		if len(val) <= jsonMaxItems {
			out := make([]interface{}, len(val))
			for i, child := range val {
				out[i] = truncateJSON(child)
			}
			return out
		}

		// Array is large: keep first jsonKeepFirst, append omitted count metadata.
		out := make([]interface{}, 0, jsonKeepFirst+1)
		for i := 0; i < jsonKeepFirst && i < len(val); i++ {
			out = append(out, truncateJSON(val[i]))
		}
		out = append(out, map[string]interface{}{
			"_omitted_items_count": len(val) - jsonKeepFirst,
		})
		return out

	default:
		return v
	}
}

// processJSON parses raw JSON, truncates arrays, and re-marshals.
func processJSON(raw []byte) ([]byte, error) {
	var root interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	truncated := truncateJSON(root)
	out, err := json.MarshalIndent(truncated, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal truncated JSON: %w", err)
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Feature B: Log & Stack Trace Pruner
// ---------------------------------------------------------------------------

var (
	rootRegex = regexp.MustCompile(`(?i)(error|exception|panic|traceback|fatal|failed)`)

	pyFrameRegex  = regexp.MustCompile(`^\s*File "(?P<path>[^"]+)", line (?P<line>\d+), in `)
	nodeFrameRegex = regexp.MustCompile(`^\s*at\s+(?:(?P<func>[^\(]+?)\s+\()?(?P<path>[^():]+):(?P<line>\d+)(?::\d+)?\)?\s*$`)
	goFrameRegex  = regexp.MustCompile(`^\s*(?P<path>[^:]+\.go):(?P<line>\d+)(?:\s+\+0x[0-9a-f]+)?$`)
	javaFrameRegex = regexp.MustCompile(`^\s*at\s+(?P<func>[^\(]+)\((?P<path>[^:]+\.java):(?P<line>\d+)\)$`)
	rustFrameRegex = regexp.MustCompile(`^\s*at\s+(?P<path>[^:]+):(?P<line>\d+)(?::\d+)?$`)
	genericFrameRegex = regexp.MustCompile(`^\s*(?:at\s+)?(?P<path>[^:]+\.(?:py|js|ts|go|java|rs)):(?P<line>\d+)(?::\d+)?.*$`)
)

func extractFrame(line string) (path, lineNo string, ok bool) {
	patterns := []*regexp.Regexp{
		pyFrameRegex, nodeFrameRegex, goFrameRegex, javaFrameRegex, rustFrameRegex, genericFrameRegex,
	}
	for _, re := range patterns {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		idx := re.SubexpIndex("path")
		if idx < 0 || idx >= len(m) {
			continue
		}
		path = m[idx]
		idxLine := re.SubexpIndex("line")
		if idxLine < 0 || idxLine >= len(m) {
			continue
		}
		lineNo = m[idxLine]
		if path != "" && lineNo != "" {
			return path, lineNo, true
		}
	}
	return "", "", false
}

func isInternalPath(path, fullLine string) bool {
	p := strings.ToLower(path)
	f := strings.ToLower(fullLine)

	builtInMarkers := []string{
		"node_modules",
		"/usr/lib",
		"/usr/local/lib",
		"site-packages",
		"dist-packages",
		".cargo/registry",
		"/rustc/",
		"/go/src/",
		"/goroot/",
		"/gopath/pkg/mod/",
		"vendor/",
		"/python3.",
		"lib/python",
		"java.base",
		"jdk.internal",
		"sun.",
		"com.sun.",
		"org.springframework.",
		"io.netty.",
		"com.google.common.",
		"kotlin.",
		"scala.",
	}

	allMarkers := append([]string{}, builtInMarkers...)
	allMarkers = append(allMarkers, customInternalMarkers...)

	for _, marker := range allMarkers {
		if strings.Contains(p, marker) || strings.Contains(f, marker) {
			return true
		}
	}
	for _, prefix := range []string{"/usr/", "/lib/", "/opt/", "/bin/", "/sbin/"} {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

func processLog(raw []byte) ([]byte, error) {
	text := string(raw)
	lines := strings.Split(text, "\n")

	stackDetected := false
	for _, line := range lines {
		if _, _, ok := extractFrame(line); ok {
			stackDetected = true
			break
		}
	}
	if !stackDetected {
		return []byte(strings.TrimSpace(text)), nil
	}

	var rootLine string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, _, ok := extractFrame(trimmed); ok {
			continue
		}
		if rootRegex.MatchString(trimmed) {
			rootLine = trimmed
		}
	}
	if rootLine == "" && len(lines) > 0 {
		rootLine = strings.TrimSpace(lines[0])
	}

	var frameLines []string
	var userFrames []string
	seenUser := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		path, lineNo, ok := extractFrame(trimmed)
		if !ok {
			continue
		}
		info := fmt.Sprintf("%s:%s", path, lineNo)
		frameLines = append(frameLines, info)

		if !isInternalPath(path, trimmed) {
			if !seenUser[info] {
				seenUser[info] = true
				userFrames = append(userFrames, info)
			}
		}
	}

	var b strings.Builder
	if rootLine != "" {
		b.WriteString("Root Error: ")
		b.WriteString(rootLine)
		b.WriteString("\n")
	}

	if len(userFrames) > 0 {
		b.WriteString("User Frames:\n")
		for _, f := range userFrames {
			b.WriteString("  - ")
			b.WriteString(f)
			b.WriteString("\n")
		}
	}

	contextStart := 0
	if len(frameLines) > 3 {
		contextStart = len(frameLines) - 3
	}
	contextLines := frameLines[contextStart:]
	b.WriteString("Context (last 3 frames):\n")
	for _, f := range contextLines {
		b.WriteString("  - ")
		b.WriteString(f)
		b.WriteString("\n")
	}

	return []byte(strings.TrimRight(b.String(), "\n")), nil
}

// ---------------------------------------------------------------------------
// Feature C: Web / HTML Content Cleaner
// ---------------------------------------------------------------------------

var (
	htmlCommentRegex   = regexp.MustCompile(`(?s)<!--.*?-->`)
	scriptStyleRegex   = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	navFooterSvgRegex  = regexp.MustCompile(`(?is)<(nav|footer|svg)[^>]*>.*?</(nav|footer|svg)>`)
	headingRegex       = regexp.MustCompile(`(?is)<h([1-3])[^>]*>(.*?)</h([1-3])>`)
	anyTagRegex        = regexp.MustCompile(`<[^>]+>`)
	base64ImageRegex   = regexp.MustCompile(`data:image/[a-zA-Z0-9.+-]+;base64,[A-Za-z0-9+/=]+`)
	whitespaceRegex    = regexp.MustCompile(`\s+`)
	headingMarkdownMap = map[string]string{"1": "#", "2": "##", "3": "###"}
)

func processHTML(raw []byte) ([]byte, error) {
	s := string(raw)

	s = htmlCommentRegex.ReplaceAllString(s, "")
	s = scriptStyleRegex.ReplaceAllString(s, "")
	s = navFooterSvgRegex.ReplaceAllString(s, "")

	s = headingRegex.ReplaceAllStringFunc(s, func(match string) string {
		sub := headingRegex.FindStringSubmatch(match)
		if len(sub) < 4 {
			return ""
		}
		level := sub[1]
		content := sub[2]
		prefix := headingMarkdownMap[level]
		return fmt.Sprintf("%s %s\n", prefix, strings.TrimSpace(content))
	})

	s = anyTagRegex.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = base64ImageRegex.ReplaceAllString(s, "")
	s = whitespaceRegex.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	return []byte(s), nil
}

// ---------------------------------------------------------------------------
// Central Processor
// ---------------------------------------------------------------------------

func process(raw []byte, mode string, session *Session) ([]byte, error) {
	var cleaned []byte
	var err error

	switch mode {
	case "json":
		cleaned, err = processJSON(raw)
	case "log":
		cleaned, err = processLog(raw)
	case "html":
		cleaned, err = processHTML(raw)
	default:
		return nil, fmt.Errorf("unsupported mode: %q", mode)
	}
	if err != nil {
		return nil, err
	}

	if session != nil {
		cleaned = session.CheckAndStore(mode, cleaned)
	}
	return cleaned, nil
}

// ---------------------------------------------------------------------------
// CLI Mode (stdin/stdout)
// ---------------------------------------------------------------------------

func runCLI() int {
	// Define flags
	mode := flag.String("mode", "", "compression mode: json|log|html|mcp")
	sessionID := flag.String("session-id", "", "session ID for persistent dedupe (optional)")
	cacheDir := flag.String("cache-dir", "", "directory for dedupe cache (default: ~/.tokencompress)")
	outputFile := flag.String("output", "", "write output to file instead of stdout")
	showVersion := flag.Bool("version", false, "print version and exit")

	// JSON truncation flags
	flag.IntVar(&jsonMaxItems, "json-max-items", 5, "maximum array length before truncation (JSON mode)")
	flag.IntVar(&jsonKeepFirst, "json-keep-first", 3, "number of items to keep when truncating (JSON mode)")

	// Custom internal path markers for log pruning (repeatable)
	flag.Func("log-internal-marker", "add custom internal path marker (repeatable)", func(s string) error {
		customInternalMarkers = append(customInternalMarkers, strings.ToLower(s))
		return nil
	})

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return 0
	}

	if *mode == "" {
		fmt.Fprintln(os.Stderr, "Error: --mode is required (json|log|html|mcp)")
		flag.Usage()
		return 2
	}

	if *mode == "mcp" {
		runMCP()
		return 0
	}

	// Determine input source: optional file argument or stdin
	var raw []byte
	var err error
	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "Error: too many input arguments; provide exactly one file or pipe via stdin")
		return 2
	}
	if len(args) == 1 {
		raw, err = os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			return 1
		}
	} else {
		raw, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			return 1
		}
	}

	// Set up session (with optional persistence)
	session := NewSession()
	if *sessionID != "" {
		cachePath := ""
		if *cacheDir != "" {
			cachePath = filepath.Join(*cacheDir, fmt.Sprintf("tokencompress_%s.json", *sessionID))
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting home dir: %v\n", err)
				return 1
			}
			cachePath = filepath.Join(home, ".tokencompress", fmt.Sprintf("cache_%s.json", *sessionID))
		}
		s, err := LoadSessionFromFile(cachePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading session cache: %v\n", err)
			return 1
		}
		session = s
		defer func() {
			if err := session.SaveSessionToFile(cachePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not save session cache: %v\n", err)
			}
		}()
	}

	out, err := process(raw, *mode, session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Write output to file or stdout
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			return 1
		}
	} else {
		fmt.Print(string(out))
		if len(out) == 0 || out[len(out)-1] != '\n' {
			fmt.Println()
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// MCP Sidecar Mode (JSON-RPC over stdio)
// ---------------------------------------------------------------------------

type mcpTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

func runMCP() {
	session := NewSession()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB buffer

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req map[string]interface{}
		if err := json.Unmarshal(line, &req); err != nil {
			resp := map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32700,
					"message": "Parse error",
				},
			}
			out, _ := json.Marshal(resp)
			fmt.Println(string(out))
			continue
		}

		method, _ := req["method"].(string)
		id, hasID := req["id"]

		if !hasID {
			continue
		}

		switch method {
		case "initialize":
			result := map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "tokencompress",
					"version": version,
				},
			}
			writeJSONRPCResult(id, result)

		case "ping":
			writeJSONRPCResult(id, map[string]interface{}{})

		case "tools/list":
			tools := []mcpTool{
				{
					Name:        "tokencompress_json",
					Description: "Truncates JSON arrays to 3 representative objects and adds _omitted_items_count.",
					InputSchema: json.RawMessage(`{"type":"object","properties":{"content":{"type":"string"}},"required":["content"]}`),
				},
				{
					Name:        "tokencompress_log",
					Description: "Prunes stack traces, filters framework boilerplate, keeps user frames and root error.",
					InputSchema: json.RawMessage(`{"type":"object","properties":{"content":{"type":"string"}},"required":["content"]}`),
				},
				{
					Name:        "tokencompress_html",
					Description: "Strips scripts, styles, nav, footer, SVG, base64 images and extracts readable text + markdown headers.",
					InputSchema: json.RawMessage(`{"type":"object","properties":{"content":{"type":"string"}},"required":["content"]}`),
				},
			}
			writeJSONRPCResult(id, map[string]interface{}{"tools": tools})

		case "tools/call":
			params, ok := req["params"].(map[string]interface{})
			if !ok {
				writeJSONRPCError(id, -32602, "Invalid params")
				continue
			}
			toolName, _ := params["name"].(string)
			arguments, _ := params["arguments"].(map[string]interface{})
			content, _ := arguments["content"].(string)

			mode := ""
			switch toolName {
			case "tokencompress_json":
				mode = "json"
			case "tokencompress_log":
				mode = "log"
			case "tokencompress_html":
				mode = "html"
			default:
				writeJSONRPCError(id, -32602, fmt.Sprintf("Unknown tool: %s", toolName))
				continue
			}

			out, err := process([]byte(content), mode, session)
			if err != nil {
				result := map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
					},
					"isError": true,
				}
				writeJSONRPCResult(id, result)
				continue
			}

			result := map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": string(out)},
				},
				"isError": false,
			}
			writeJSONRPCResult(id, result)

		default:
			writeJSONRPCError(id, -32601, "Method not found")
		}
	}
}

func writeJSONRPCResult(id interface{}, result interface{}) {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}

func writeJSONRPCError(id interface{}, code int, message string) {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	out, _ := json.Marshal(resp)
	fmt.Println(string(out))
}

func main() {
	os.Exit(runCLI())
}
