// main.go – Anaesthetic Billing Checker server
//
// Replaces the Node.js web-server and MBS update script with a single
// Go binary.  Serves the static frontend from app/ and provides API
// endpoints for MBS catalog update and status.
//
// Copyright (c) 2026 Bernard McClement
// Licensed under the MIT License. See LICENSE file in the project root.
// You may copy, modify, and redistribute this software with attribution.

package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ── Configuration ─────────────────────────────────────────────────

const (
	defaultPort    = 8080
	baseURL        = "https://www.mbsonline.gov.au"
	downloadsIndex = baseURL + "/internet/mbsonline/publishing.nsf/Content/downloads"
	userAgent      = "AnaestheticBillingChecker/1.0"
)

// ── Global state ──────────────────────────────────────────────────

var (
	rootDir     string
	appDir      string
	mbsDataFile string

	updateMu       sync.Mutex
	updateProgress bool
	firstOpenDone  bool
	firstOpenMu    sync.Mutex
)

// ── Main ──────────────────────────────────────────────────────────

func main() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal("Cannot determine executable path:", err)
	}
	rootDir = filepath.Dir(exe)

	// If the binary is run from a "go run" temp dir, fall back to cwd
	if _, err := os.Stat(filepath.Join(rootDir, "app")); os.IsNotExist(err) {
		rootDir, _ = os.Getwd()
	}

	appDir = filepath.Join(rootDir, "app")
	mbsDataFile = filepath.Join(appDir, "mbs-data.js")

	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/meta", handleMeta)
	mux.HandleFunc("/api/update-mbs", handleUpdateMBS)

	// Static file serving
	mux.HandleFunc("/", handleStatic)

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute, // long for MBS update
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down…")
		srv.Close()
	}()

	fmt.Println()
	fmt.Println("  Anaesthetic Billing Checker")
	fmt.Printf("  Running at  http://localhost:%d\n", port)
	fmt.Println("  First page open triggers automatic MBS data update.")
	fmt.Println()

	// Open browser after short delay
	go func() {
		time.Sleep(1500 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://localhost:%d", port))
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// ── HTTP Handlers ─────────────────────────────────────────────────

func handleMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		textResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	triggerFirstOpenUpdate()
	meta := readMeta()
	updateMu.Lock()
	inProg := updateProgress
	updateMu.Unlock()
	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":               true,
		"meta":             meta,
		"updateInProgress": inProg,
	})
}

func handleUpdateMBS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		textResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	updateMu.Lock()
	if updateProgress {
		updateMu.Unlock()
		jsonResponse(w, http.StatusConflict, map[string]any{
			"ok":    false,
			"error": "Update already in progress.",
		})
		return
	}
	updateProgress = true
	updateMu.Unlock()

	err := runMBSUpdate()

	updateMu.Lock()
	updateProgress = false
	updateMu.Unlock()

	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	meta := readMeta()
	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":   true,
		"meta": meta,
	})
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		textResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Trigger first-open update asynchronously
	go triggerFirstOpenUpdate()

	p := r.URL.Path
	if p == "/" || p == "/index.html" {
		p = "/index.html"
	}

	// Sanitise
	cleaned := strings.TrimLeft(p, "/")
	if strings.Contains(cleaned, "..") {
		textResponse(w, http.StatusBadRequest, "Bad request")
		return
	}
	if cleaned == "" {
		cleaned = "index.html"
	}

	filePath := filepath.Join(appDir, filepath.FromSlash(cleaned))
	data, err := os.ReadFile(filePath)
	if err != nil {
		textResponse(w, http.StatusNotFound, "Not found")
		return
	}

	ct := mimeForExt(filepath.Ext(filePath))
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ── First-open auto-update ────────────────────────────────────────

func triggerFirstOpenUpdate() {
	firstOpenMu.Lock()
	if firstOpenDone {
		firstOpenMu.Unlock()
		return
	}
	firstOpenDone = true
	firstOpenMu.Unlock()

	if os.Getenv("DISABLE_AUTO_UPDATE") == "1" {
		return
	}

	updateMu.Lock()
	if updateProgress {
		updateMu.Unlock()
		return
	}
	updateProgress = true
	updateMu.Unlock()

	err := runMBSUpdate()

	updateMu.Lock()
	updateProgress = false
	updateMu.Unlock()

	if err != nil {
		log.Println("Auto-update failed:", err)
	}
}

// ── MBS Update Logic ─────────────────────────────────────────────
// This is the Go port of update-mbs-data.mjs.

func runMBSUpdate() error {
	log.Println("Fetching MBS downloads index…")
	indexHTML, err := fetchText(downloadsIndex)
	if err != nil {
		return fmt.Errorf("fetching downloads index: %w", err)
	}

	latestPageURL := findLatestDownloadsPage(indexHTML)
	if latestPageURL == "" {
		return fmt.Errorf("could not locate latest downloads page on MBS Online")
	}
	log.Println("Latest downloads page:", latestPageURL)

	pageHTML, err := fetchText(latestPageURL)
	if err != nil {
		return fmt.Errorf("fetching downloads page: %w", err)
	}

	xmlURL := findXMLDownloadURL(pageHTML)
	if xmlURL == "" {
		return fmt.Errorf("no XML download link found on %s", latestPageURL)
	}
	log.Println("XML download:", xmlURL)

	log.Println("Downloading XML (this may take a moment)…")
	xmlText, err := fetchText(xmlURL)
	if err != nil {
		return fmt.Errorf("downloading XML: %w", err)
	}
	log.Printf("XML size: %d KB\n", len(xmlText)/1024)

	items, notes := parseXML(xmlText)
	totalItems := len(items)
	log.Printf("Parsed %d MBS items.\n", totalItems)

	meta := map[string]any{
		"source":            "MBS Online official XML",
		"downloadsIndexUrl": downloadsIndex,
		"downloadsPageUrl":  latestPageURL,
		"xmlDownloadUrl":    xmlURL,
		"generatedAt":       time.Now().UTC().Format(time.RFC3339),
		"totalItems":        totalItems,
		"parserNotes":       notes,
	}

	itemsJSON, _ := json.MarshalIndent(items, "", "  ")
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")

	js := fmt.Sprintf(
		"// Auto-generated by billing-checker – do not edit\nwindow.MBS_DATA = %s;\nwindow.MBS_DATA_META = %s;\n",
		string(itemsJSON), string(metaJSON),
	)

	if err := os.MkdirAll(filepath.Dir(mbsDataFile), 0o755); err != nil {
		return fmt.Errorf("creating app dir: %w", err)
	}
	if err := os.WriteFile(mbsDataFile, []byte(js), 0o644); err != nil {
		return fmt.Errorf("writing mbs-data.js: %w", err)
	}

	log.Printf("Wrote %s (%d items)\n", mbsDataFile, totalItems)
	return nil
}

// ── HTTP fetch helper ─────────────────────────────────────────────

func fetchText(rawURL string) (string, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, rawURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ── Link extraction helpers ───────────────────────────────────────

var hrefRe = regexp.MustCompile(`href\s*=\s*"([^"]+)"`)
var downloadsPageRe = regexp.MustCompile(`(?i)/Content/Downloads-`)
var xmlFileRe = regexp.MustCompile(`(?i)\$FILE/.*\.XML$`)
var dateYYYYMMDDRe = regexp.MustCompile(`(20\d{6})`)
var date6Re = regexp.MustCompile(`\b(\d{6})\b`)
var dateYM = regexp.MustCompile(`(20\d{2})[-_](\d{2})`)
var versionRe = regexp.MustCompile(`(?i)version\s*(\d+)`)

func extractHrefs(html string) []string {
	matches := hrefRe.FindAllStringSubmatch(html, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

func toAbsolute(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return baseURL + href
	}
	return baseURL + "/" + href
}

func dateScore(text string) int {
	decoded, _ := url.QueryUnescape(text)
	if m := dateYYYYMMDDRe.FindString(decoded); m != "" {
		n, _ := strconv.Atoi(m)
		return n
	}
	if m := date6Re.FindStringSubmatch(decoded); m != nil {
		r := m[1]
		n, _ := strconv.Atoi("20" + r[:2] + r[2:4] + r[4:6])
		return n
	}
	if m := dateYM.FindStringSubmatch(decoded); m != nil {
		n, _ := strconv.Atoi(m[1] + m[2] + "01")
		return n
	}
	return 0
}

func versionScore(text string) int {
	decoded, _ := url.QueryUnescape(text)
	if m := versionRe.FindStringSubmatch(decoded); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 1
}

func findLatestDownloadsPage(html string) string {
	type candidate struct {
		url   string
		score int
	}
	var candidates []candidate
	for _, href := range extractHrefs(html) {
		abs := toAbsolute(href)
		if downloadsPageRe.MatchString(abs) {
			candidates = append(candidates, candidate{abs, dateScore(abs)})
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	return candidates[0].url
}

func findXMLDownloadURL(html string) string {
	type candidate struct {
		url string
		d   int
		v   int
	}
	var candidates []candidate
	for _, href := range extractHrefs(html) {
		abs := toAbsolute(href)
		if xmlFileRe.MatchString(abs) {
			candidates = append(candidates, candidate{abs, dateScore(abs), versionScore(abs)})
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].d != candidates[j].d {
			return candidates[i].d > candidates[j].d
		}
		return candidates[i].v > candidates[j].v
	})
	return candidates[0].url
}

// ── XML Parsing ───────────────────────────────────────────────────
// The MBS XML has a deeply nested structure. We use a SAX-like
// approach to walk all elements and extract item-number/description/fee
// triples, matching the logic from the JS version.

// MBSItem represents a parsed MBS schedule item.
type MBSItem struct {
	Description string   `json:"description"`
	ScheduleFee *float64 `json:"scheduleFee"`
}

func parseXML(xmlText string) (map[string]*MBSItem, []string) {
	items := make(map[string]*MBSItem)
	notes := []string{"Descriptions and schedule fees extracted from latest official XML."}

	// Use generic XML walking – decode into a tree of maps
	decoder := xml.NewDecoder(strings.NewReader(xmlText))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose

	// We'll parse the XML element by element, collecting field name/value
	// pairs per element and looking for item codes, descriptions, and fees.
	type flatNode struct {
		fields map[string]string
	}

	// Recursive walk using a stack-based approach
	var stack []map[string]string
	stack = append(stack, make(map[string]string))

	var currentTag string

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			newNode := make(map[string]string)
			// Copy attributes as fields
			for _, attr := range t.Attr {
				newNode[attr.Name.Local] = strings.TrimSpace(attr.Value)
			}
			stack = append(stack, newNode)
			currentTag = t.Name.Local

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				node := stack[len(stack)-1]
				node[currentTag] = text
			}

		case xml.EndElement:
			if len(stack) < 2 {
				continue
			}
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Check if this node has item code candidates
			codes := codeCandidates(node)
			if len(codes) == 0 {
				// propagate fields up to parent
				parent := stack[len(stack)-1]
				for k, v := range node {
					if _, exists := parent[k]; !exists {
						parent[k] = v
					}
				}
				continue
			}

			desc := bestDescription(node)
			fee := bestFee(node)
			if desc == "" {
				continue
			}

			for _, code := range codes {
				prev, exists := items[code]
				if !exists {
					items[code] = &MBSItem{Description: desc, ScheduleFee: fee}
				} else {
					prev.Description = pickLonger(prev.Description, desc)
					if prev.ScheduleFee == nil && fee != nil {
						prev.ScheduleFee = fee
					}
				}
			}
		}
	}

	return items, notes
}

var codeFieldRe = regexp.MustCompile(`(?i)(item|code|num|number)`)
var fiveDigitRe = regexp.MustCompile(`^\d{5}$`)
var descFieldRe = regexp.MustCompile(`(?i)(descr|description|itemtext|scheduletext|service|title|display)`)
var feeFieldRe = regexp.MustCompile(`(?i)(schedule.?fee|fee|benefit|amt|amount|dollar)`)

func codeCandidates(fields map[string]string) []string {
	seen := make(map[string]bool)
	var out []string
	for k, v := range fields {
		if codeFieldRe.MatchString(k) && fiveDigitRe.MatchString(v) {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	return out
}

func bestDescription(fields map[string]string) string {
	var best string
	for k, v := range fields {
		if descFieldRe.MatchString(k) && len(v) >= 12 && !fiveDigitRe.MatchString(v) {
			if len(v) > len(best) {
				best = v
			}
		}
	}
	return best
}

func bestFee(fields map[string]string) *float64 {
	var best *float64
	for k, v := range fields {
		if feeFieldRe.MatchString(k) {
			f := parseMoney(v)
			if f > 0 && f < 50000 {
				if best == nil || f > *best {
					val := f
					best = &val
				}
			}
		}
	}
	return best
}

func parseMoney(s string) float64 {
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			return r
		}
		return -1
	}, s)
	if cleaned == "" {
		return 0
	}
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return f
}

func pickLonger(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if len(b) > len(a) {
		return b
	}
	return a
}

// ── Read catalog metadata from generated JS ───────────────────────

var metaRe = regexp.MustCompile(`window\.MBS_DATA_META\s*=\s*(\{[\s\S]*?\});`)

func readMeta() map[string]any {
	fallback := map[string]any{
		"source":      "not-loaded",
		"generatedAt": nil,
		"totalItems":  0,
	}
	data, err := os.ReadFile(mbsDataFile)
	if err != nil {
		return fallback
	}
	m := metaRe.FindSubmatch(data)
	if m == nil {
		return fallback
	}
	var result map[string]any
	if err := json.Unmarshal(m[1], &result); err != nil {
		return fallback
	}
	return result
}

// ── MIME types ─────────────────────────────────────────────────────

func mimeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

// ── Response helpers ──────────────────────────────────────────────

func jsonResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func textResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(body))
}

// ── Browser open ──────────────────────────────────────────────────

func openBrowser(url string) {
	// Only auto-open if we can actually bind to the port
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", defaultPort), 500*time.Millisecond)
	if err != nil {
		return
	}
	conn.Close()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run()
}
