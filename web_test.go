package main

import (
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestWebDashboardShowsRankingsReasonsAndMethodology(t *testing.T) {
	handler := NewHandler()
	response := webRequest(t, handler, http.MethodGet, "/")

	if response.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", got)
	}

	body := response.Body.String()
	visibleBody := html.UnescapeString(body)
	for _, label := range []string{
		"Lead portfolio", "Search leads", "Priority", "Rank", "Company", "Score",
		"Scoring reasons", "Methodology", "fictional", "advisory",
		"30 points", "25 points", "10 points", "5 points", "100 points",
	} {
		if !strings.Contains(strings.ToLower(body), strings.ToLower(label)) {
			t.Errorf("dashboard is missing visible label or explanation %q", label)
		}
	}

	scores := webRankedScores(t)
	if got := leadRowCount(body); got != len(scores) {
		t.Errorf("rendered lead rows = %d, want full ranking of %d", got, len(scores))
	}
	lastPosition := -1
	for _, score := range scores {
		name := html.EscapeString(score.Lead.Name)
		position := strings.Index(body, name)
		if position < 0 {
			t.Errorf("dashboard does not show ranked lead %q", score.Lead.Name)
		} else if position <= lastPosition {
			t.Errorf("dashboard lead %q is not in ranking order", score.Lead.Name)
		}
		lastPosition = position
		for _, reason := range score.Reasons {
			if !strings.Contains(visibleBody, reason) {
				t.Errorf("dashboard does not show reason %q for lead %q", reason, score.Lead.Name)
			}
		}
		for _, missing := range score.Missing {
			if !strings.Contains(visibleBody, missing) {
				t.Errorf("dashboard does not explain missing data %q for lead %q", missing, score.Lead.Name)
			}
		}
	}
}

func TestWebRejectsUnknownBand(t *testing.T) {
	response := webRequest(t, NewHandler(), http.MethodGet, "/?band=urgent")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("GET /?band=urgent status = %d, want %d; body: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
	const want = "Invalid priority band. Choose high, medium, low, or omit the band filter.\n"
	if got := response.Body.String(); got != want {
		t.Errorf("GET /?band=urgent body = %q, want %q", got, want)
	}
	if got := leadRowCount(response.Body.String()); got != 0 {
		t.Errorf("invalid band rendered %d lead rows, want none", got)
	}
}

func TestWebBandValidationBoundaries(t *testing.T) {
	scores := webRankedScores(t)
	cases := []struct {
		name, target, band string
		invalid            bool
	}{
		{name: "omitted", target: "/"},
		{name: "empty", target: "/?band="},
		{name: "high", target: "/?band=high", band: "high"},
		{name: "medium", target: "/?band=medium", band: "medium"},
		{name: "low", target: "/?band=low", band: "low"},
		{name: "whitespace-only invalid", target: "/?band=%20", invalid: true},
		{name: "later duplicate invalid", target: "/?band=low&band=urgent", invalid: true},
	}
	const recovery = "Invalid priority band. Choose high, medium, low, or omit the band filter.\n"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := webRequest(t, NewHandler(), http.MethodGet, tc.target)
			if tc.invalid {
				if response.Code != http.StatusBadRequest {
					t.Fatalf("GET %s status = %d, want %d", tc.target, response.Code, http.StatusBadRequest)
				}
				if got := response.Body.String(); got != recovery {
					t.Errorf("GET %s body = %q, want %q", tc.target, got, recovery)
				}
				if got := leadRowCount(response.Body.String()); got != 0 {
					t.Errorf("invalid band rendered %d lead rows, want none", got)
				}
				return
			}

			if response.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", tc.target, response.Code, http.StatusOK)
			}
			wantRows := 0
			for _, score := range scores {
				if tc.band == "" || score.Band == tc.band {
					wantRows++
				}
			}
			if got := leadRowCount(response.Body.String()); got != wantRows {
				t.Errorf("GET %s rendered %d lead rows, want %d", tc.target, got, wantRows)
			}
		})
	}
}

func TestWebSearchAndBandFiltersWorkTogether(t *testing.T) {
	scores := webRankedScores(t)
	selected := uniqueSearchLead(t, scores)
	query := url.QueryEscape(strings.ToUpper(selected.Lead.Name))
	target := "/?q=" + query + "&band=" + url.QueryEscape(selected.Band)
	response := webRequest(t, NewHandler(), http.MethodGet, target)

	if response.Code != http.StatusOK {
		t.Fatalf("filtered GET status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := leadRowCount(response.Body.String()); got != 1 {
		t.Fatalf("combined search and band filters show %d rows, want 1", got)
	}
	if !strings.Contains(response.Body.String(), html.EscapeString(selected.Lead.Name)) {
		t.Errorf("combined filters omitted matching lead %q", selected.Lead.Name)
	}
	for _, otherBand := range []string{"high", "medium", "low"} {
		if otherBand == selected.Band {
			continue
		}
		mismatch := webRequest(t, NewHandler(), http.MethodGet, "/?q="+query+"&band="+otherBand)
		if got := leadRowCount(mismatch.Body.String()); got != 0 {
			t.Errorf("search for %q with non-matching band %q rendered %d rows, want 0", selected.Lead.Name, otherBand, got)
		}
		break
	}
	if !strings.Contains(response.Body.String(), "Clear filters") {
		t.Error("filtered dashboard has no clear-filters control")
	}
}

func TestWebNoResultsAndReset(t *testing.T) {
	handler := NewHandler()
	scores := webRankedScores(t)
	selected := uniqueSearchLead(t, scores)

	filtered := webRequest(t, handler, http.MethodGet, "/?q="+url.QueryEscape(selected.Lead.Name)+"&band="+url.QueryEscape(selected.Band))
	if got := leadRowCount(filtered.Body.String()); got != 1 {
		t.Fatalf("setup filter shows %d rows, want 1", got)
	}
	if !strings.Contains(filtered.Body.String(), "href=\"/\"") {
		t.Fatal("clear-filters control does not link to the unfiltered dashboard")
	}

	reset := webRequest(t, handler, http.MethodGet, "/")
	if got := leadRowCount(reset.Body.String()); got != len(scores) {
		t.Errorf("reset dashboard shows %d rows, want all %d", got, len(scores))
	}

	empty := webRequest(t, handler, http.MethodGet, "/?q="+url.QueryEscape("no-lead-can-match-this-query-847263"))
	if got := leadRowCount(empty.Body.String()); got != 0 {
		t.Errorf("no-results search rendered %d lead rows, want 0", got)
	}
	if !strings.Contains(strings.ToLower(empty.Body.String()), "no leads match") {
		t.Error("no-results state is not explained to the user")
	}
}

func TestWebSearchInputIsEscaped(t *testing.T) {
	query := `<script>alert("x")</script>&`
	response := webRequest(t, NewHandler(), http.MethodGet, "/?q="+url.QueryEscape(query))
	body := response.Body.String()

	if strings.Contains(body, query) || strings.Contains(body, "<script>") {
		t.Fatal("search input was rendered as unescaped HTML")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Error("escaped search input is not preserved in the search field")
	}
}

func TestWebBandValidRepeatedValuesKeepFirstNonEmpty(t *testing.T) {
	scores := webRankedScores(t)
	cases := []struct {
		name, target, wantBand string
	}{
		{name: "low before high", target: "/?band=low&band=high", wantBand: "low"},
		{name: "high before low", target: "/?band=high&band=low", wantBand: "high"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := webRequest(t, NewHandler(), http.MethodGet, tc.target)
			if response.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", tc.target, response.Code, http.StatusOK)
			}

			body := response.Body.String()
			wantRows := 0
			for _, score := range scores {
				want := score.Band == tc.wantBand
				if want {
					wantRows++
				}
				if got := strings.Contains(body, html.EscapeString(score.Lead.Name)); got != want {
					t.Errorf("GET %s includes lead %q = %t, want %t", tc.target, score.Lead.Name, got, want)
				}
			}
			if wantRows == 0 {
				t.Fatalf("test fixtures contain no %q leads", tc.wantBand)
			}
			if got := leadRowCount(body); got != wantRows {
				t.Errorf("GET %s rendered %d rows, want %d", tc.target, got, wantRows)
			}
		})
	}
}

func TestWebAPIIsFullSeedRankingJSON(t *testing.T) {
	want := webRankedScores(t)
	target := "/api/leads?q=does-not-match&band=low"
	response := webRequest(t, NewHandler(), http.MethodGet, target)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/leads status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var got []Score
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode API response: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("API ranking does not equal Rank(SampleLeads())\n got: %#v\nwant: %#v", got, want)
	}
	if len(got) > 0 {
		var objects []map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &objects); err != nil {
			t.Fatalf("decode API fields: %v", err)
		}
		for _, key := range []string{"lead", "score", "band", "reasons", "missing", "version"} {
			if _, ok := objects[0][key]; !ok {
				t.Errorf("API object lacks snake_case field %q", key)
			}
		}
		var leadFields map[string]json.RawMessage
		if err := json.Unmarshal(objects[0]["lead"], &leadFields); err != nil {
			t.Fatalf("decode API lead fields: %v", err)
		}
		for _, key := range []string{"id", "name", "company", "requested_demo", "budget_confirmed", "timeline_days", "interactions"} {
			if _, ok := leadFields[key]; !ok {
				t.Errorf("API lead lacks snake_case field %q", key)
			}
		}
	}
}

func TestWebStaticAssetAndMethodHandling(t *testing.T) {
	handler := NewHandler()
	asset := webRequest(t, handler, http.MethodGet, "/style.css")
	if asset.Code != http.StatusOK {
		t.Fatalf("GET stylesheet status = %d, want %d", asset.Code, http.StatusOK)
	}
	if got := asset.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/css") {
		t.Errorf("stylesheet Content-Type = %q, want text/css", got)
	}
	if asset.Body.Len() == 0 {
		t.Error("embedded stylesheet is empty")
	}

	for _, path := range []string{"/", "/api/leads", "/style.css"} {
		for _, method := range []string{http.MethodPost, http.MethodHead} {
			t.Run(method+" "+path, func(t *testing.T) {
				response := webRequest(t, handler, method, path)
				if response.Code != http.StatusMethodNotAllowed {
					t.Errorf("%s %s status = %d, want %d", method, path, response.Code, http.StatusMethodNotAllowed)
				}
				if got := response.Header().Get("Allow"); got != http.MethodGet {
					t.Errorf("%s %s Allow = %q, want %q", method, path, got, http.MethodGet)
				}
			})
		}
	}

	unknown := webRequest(t, handler, http.MethodGet, "/not-a-route")
	if unknown.Code != http.StatusNotFound {
		t.Errorf("GET unknown route status = %d, want %d", unknown.Code, http.StatusNotFound)
	}
	unknownMethod := webRequest(t, handler, http.MethodPost, "/not-a-route")
	if unknownMethod.Code != http.StatusNotFound {
		t.Errorf("POST unknown route status = %d, want %d", unknownMethod.Code, http.StatusNotFound)
	}
}

func webRequest(t *testing.T, handler http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func webRankedScores(t *testing.T) []Score {
	t.Helper()
	scores, err := Rank(SampleLeads())
	if err != nil {
		t.Fatalf("Rank(SampleLeads()): %v", err)
	}
	if len(scores) == 0 {
		t.Fatal("SampleLeads() returned no leads")
	}
	return scores
}

func uniqueSearchLead(t *testing.T, scores []Score) Score {
	t.Helper()
	bandCounts := make(map[string]int)
	for _, score := range scores {
		bandCounts[score.Band]++
	}
	for _, candidate := range scores {
		query := strings.ToLower(candidate.Lead.Name)
		matches := 0
		for _, score := range scores {
			if strings.Contains(strings.ToLower(score.Lead.Name), query) || strings.Contains(strings.ToLower(score.Lead.Company), query) {
				matches++
			}
		}
		if matches == 1 && bandCounts[candidate.Band] > 1 {
			return candidate
		}
	}
	t.Fatal("sample ranking has no uniquely searchable lead in a shared priority band")
	return Score{}
}

func leadRowCount(body string) int {
	return strings.Count(body, `class="lead-row"`)
}
