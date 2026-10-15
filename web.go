package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web/index.html web/style.css
var webAssets embed.FS

var dashboardTemplate = template.Must(template.ParseFS(webAssets, "web/index.html"))

type portfolioHandler struct {
	ranked  []Score
	rankErr error
}

type dashboardLead struct {
	Rank  int
	Score Score
}

type dashboardPage struct {
	Query   string
	Band    string
	Leads   []dashboardLead
	Total   int
	Version string
}

// NewHandler returns the read-only portfolio dashboard and JSON ranking API.
func NewHandler() http.Handler {
	ranked, err := Rank(SampleLeads())
	return &portfolioHandler{ranked: ranked, rankErr: err}
}

func (h *portfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.serveDashboard(w, r)
	case "/api/leads":
		h.serveAPI(w, r)
	case "/style.css":
		h.serveStylesheet(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *portfolioHandler) serveDashboard(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	if h.rankErr != nil {
		http.Error(w, "The sample lead ranking could not be calculated.", http.StatusInternalServerError)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	band, valid := validBand(r.URL.Query()["band"])
	if !valid {
		http.Error(w, "Invalid priority band. Choose high, medium, low, or omit the band filter.", http.StatusBadRequest)
		return
	}
	queryLower := strings.ToLower(query)
	filtered := make([]dashboardLead, 0, len(h.ranked))
	for _, score := range h.ranked {
		if band != "" && score.Band != band {
			continue
		}
		if queryLower != "" && !matchesLead(score.Lead, queryLower) {
			continue
		}
		filtered = append(filtered, dashboardLead{Rank: len(filtered) + 1, Score: score})
	}

	version := ""
	if len(h.ranked) > 0 {
		version = h.ranked[0].Version
	}
	page := dashboardPage{
		Query:   query,
		Band:    band,
		Leads:   filtered,
		Total:   len(h.ranked),
		Version: version,
	}
	var rendered bytes.Buffer
	if err := dashboardTemplate.ExecuteTemplate(&rendered, "index.html", page); err != nil {
		http.Error(w, "The lead portfolio could not be rendered.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(rendered.Bytes())
}

func (h *portfolioHandler) serveAPI(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	if h.rankErr != nil {
		http.Error(w, "The sample lead ranking could not be calculated.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_ = json.NewEncoder(w).Encode(h.ranked)
}

func (h *portfolioHandler) serveStylesheet(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	style, err := fs.ReadFile(webAssets, "web/style.css")
	if err != nil {
		http.Error(w, "The stylesheet is unavailable.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(style)
}

func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}
	w.Header().Set("Allow", http.MethodGet)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func validBand(values []string) (string, bool) {
	band := ""
	for _, value := range values {
		if value == "" {
			continue
		}
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			return "", false
		}
		switch normalized {
		case "high", "medium", "low":
			if band == "" {
				band = normalized
			}
		default:
			return "", false
		}
	}
	return band, true
}

func matchesLead(lead Lead, query string) bool {
	return strings.Contains(strings.ToLower(lead.Name), query) ||
		strings.Contains(strings.ToLower(lead.Company), query) ||
		strings.Contains(strings.ToLower(lead.ID), query)
}
