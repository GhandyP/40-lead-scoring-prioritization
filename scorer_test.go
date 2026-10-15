package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRankScoresAndBands(t *testing.T) {
	intValue := func(value int) *int { return &value }
	newLead := func(id string, demo, budget bool, timeline, interactions *int) Lead {
		return Lead{
			ID:              id,
			RequestedDemo:   demo,
			BudgetConfirmed: budget,
			TimelineDays:    timeline,
			Interactions:    interactions,
		}
	}

	tests := []struct {
		name       string
		lead       Lead
		wantScore  int
		wantBand   string
		wantReason []string
		wantAbsent []string
	}{
		{
			name:      "all scoring features reach the maximum",
			lead:      newLead("all", true, true, intValue(30), intValue(7)),
			wantScore: 100,
			wantBand:  "high",
			wantReason: []string{
				"Demo requested (+30)",
				"Budget confirmed (+25)",
				"Timeline within 30 days (+25)",
				"Interactions: 4 or more (+20; capped at 4)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "high band begins at 70",
			lead:      newLead("high-boundary", false, true, intValue(30), intValue(4)),
			wantScore: 70,
			wantBand:  "high",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget confirmed (+25)",
				"Timeline within 30 days (+25)",
				"Interactions: 4 or more (+20; capped at 4)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "medium band upper neighbor is 65",
			lead:      newLead("medium-upper", true, true, intValue(90), intValue(0)),
			wantScore: 65,
			wantBand:  "medium",
			wantReason: []string{
				"Demo requested (+30)",
				"Budget confirmed (+25)",
				"Timeline 31–90 days (+10)",
				"Interactions: 0 (+0)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "medium band begins at 40",
			lead:      newLead("medium-boundary", true, false, nil, intValue(2)),
			wantScore: 40,
			wantBand:  "medium",
			wantReason: []string{
				"Demo requested (+30)",
				"Budget not confirmed (+0)",
				"Timeline unavailable (missing)",
				"Interactions: 2 (+10)",
			},
			wantAbsent: []string{"timeline_days"},
		},
		{
			name:      "low band upper neighbor is 35",
			lead:      newLead("low-upper", true, false, nil, intValue(1)),
			wantScore: 35,
			wantBand:  "low",
			wantReason: []string{
				"Demo requested (+30)",
				"Budget not confirmed (+0)",
				"Timeline unavailable (missing)",
				"Interactions: 1 (+5)",
			},
			wantAbsent: []string{"timeline_days"},
		},
		{
			name:      "timeline starts earning ten points at 31 days",
			lead:      newLead("timeline-31", false, false, intValue(31), intValue(0)),
			wantScore: 10,
			wantBand:  "low",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget not confirmed (+0)",
				"Timeline 31–90 days (+10)",
				"Interactions: 0 (+0)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "timeline earns ten points through 90 days",
			lead:      newLead("timeline-90", false, false, intValue(90), intValue(0)),
			wantScore: 10,
			wantBand:  "low",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget not confirmed (+0)",
				"Timeline 31–90 days (+10)",
				"Interactions: 0 (+0)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "timeline after 90 days earns no points",
			lead:      newLead("timeline-91", false, false, intValue(91), intValue(0)),
			wantScore: 0,
			wantBand:  "low",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget not confirmed (+0)",
				"Timeline over 90 days (+0)",
				"Interactions: 0 (+0)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "explicit numeric zero is present data",
			lead:      newLead("zero", false, false, intValue(0), intValue(0)),
			wantScore: 25,
			wantBand:  "low",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget not confirmed (+0)",
				"Timeline within 30 days (+25)",
				"Interactions: 0 (+0)",
			},
			wantAbsent: []string{},
		},
		{
			name:      "nil numeric values are missing rather than zero",
			lead:      newLead("missing", false, false, nil, nil),
			wantScore: 0,
			wantBand:  "low",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget not confirmed (+0)",
				"Timeline unavailable (missing)",
				"Interactions unavailable (missing)",
			},
			wantAbsent: []string{"timeline_days", "interactions"},
		},
		{
			name:      "interactions are capped at four",
			lead:      newLead("clamped", false, false, intValue(91), intValue(9)),
			wantScore: 20,
			wantBand:  "low",
			wantReason: []string{
				"Demo not requested (+0)",
				"Budget not confirmed (+0)",
				"Timeline over 90 days (+0)",
				"Interactions: 4 or more (+20; capped at 4)",
			},
			wantAbsent: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Rank([]Lead{test.lead})
			if err != nil {
				t.Fatalf("Rank() error = %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("Rank() returned %d results, want 1", len(got))
			}
			if got[0].Score != test.wantScore {
				t.Errorf("Score = %d, want %d", got[0].Score, test.wantScore)
			}
			if got[0].Band != test.wantBand {
				t.Errorf("Band = %q, want %q", got[0].Band, test.wantBand)
			}
			if got[0].Version != "v1" {
				t.Errorf("Version = %q, want %q", got[0].Version, "v1")
			}
			if !reflect.DeepEqual(got[0].Reasons, test.wantReason) {
				t.Errorf("Reasons = %#v, want %#v", got[0].Reasons, test.wantReason)
			}
			if !reflect.DeepEqual(got[0].Missing, test.wantAbsent) {
				t.Errorf("Missing = %#v, want %#v", got[0].Missing, test.wantAbsent)
			}
		})
	}
}

func TestRankOrdersByScoreThenIDAndIsDeterministic(t *testing.T) {
	intValue := func(value int) *int { return &value }
	leads := []Lead{
		{ID: "tie-z", RequestedDemo: true},
		{ID: "low"},
		{ID: "high", RequestedDemo: true, TimelineDays: intValue(30)},
		{ID: "tie-a", RequestedDemo: true},
	}

	first, err := Rank(leads)
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}
	second, err := Rank(leads)
	if err != nil {
		t.Fatalf("second Rank() error = %v", err)
	}

	wantIDs := []string{"high", "tie-a", "tie-z", "low"}
	gotIDs := make([]string, len(first))
	for i, score := range first {
		gotIDs[i] = score.Lead.ID
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("ranked IDs = %#v, want %#v", gotIDs, wantIDs)
	}
	if !reflect.DeepEqual(first, second) {
		t.Errorf("repeated Rank() results differ:\nfirst:  %#v\nsecond: %#v", first, second)
	}
}

func TestRankDoesNotModifyInput(t *testing.T) {
	intValue := func(value int) *int { return &value }
	leads := []Lead{
		{ID: "z", TimelineDays: intValue(31), Interactions: intValue(2)},
		{ID: "a", RequestedDemo: true, TimelineDays: intValue(7)},
	}
	before := cloneLeadsForTest(leads)

	if _, err := Rank(leads); err != nil {
		t.Fatalf("Rank() error = %v", err)
	}
	if !reflect.DeepEqual(leads, before) {
		t.Errorf("Rank() modified input:\ngot:  %#v\nwant: %#v", leads, before)
	}
}

func TestRankRejectsInvalidLeadsWithoutMutation(t *testing.T) {
	intValue := func(value int) *int { return &value }
	tests := []struct {
		name  string
		leads []Lead
	}{
		{name: "blank ID", leads: []Lead{{ID: ""}}},
		{name: "whitespace-only ID", leads: []Lead{{ID: " \t "}}},
		{name: "negative timeline", leads: []Lead{{ID: "negative-timeline", TimelineDays: intValue(-1)}}},
		{name: "negative interactions", leads: []Lead{{ID: "negative-interactions", Interactions: intValue(-1)}}},
		{name: "duplicate IDs", leads: []Lead{{ID: "duplicate"}, {ID: "duplicate"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := cloneLeadsForTest(test.leads)
			if _, err := Rank(test.leads); err == nil {
				t.Fatal("Rank() error = nil, want validation error")
			}
			if !reflect.DeepEqual(test.leads, before) {
				t.Errorf("Rank() modified invalid input:\ngot:  %#v\nwant: %#v", test.leads, before)
			}
		})
	}
}

func TestRankJSONUsesSnakeCase(t *testing.T) {
	intValue := func(value int) *int { return &value }
	got, err := Rank([]Lead{{
		ID:              "json",
		RequestedDemo:   true,
		BudgetConfirmed: true,
		TimelineDays:    intValue(30),
		Interactions:    intValue(4),
	}})
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}
	encoded, err := json.Marshal(got[0])
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var score map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &score); err != nil {
		t.Fatalf("json.Unmarshal(score) error = %v", err)
	}
	for _, key := range []string{"lead", "score", "band", "reasons", "missing", "version"} {
		if _, ok := score[key]; !ok {
			t.Errorf("score JSON is missing key %q: %s", key, encoded)
		}
	}
	var lead map[string]json.RawMessage
	if err := json.Unmarshal(score["lead"], &lead); err != nil {
		t.Fatalf("json.Unmarshal(lead) error = %v", err)
	}
	for _, key := range []string{"id", "requested_demo", "budget_confirmed", "timeline_days", "interactions"} {
		if _, ok := lead[key]; !ok {
			t.Errorf("lead JSON is missing key %q: %s", key, score["lead"])
		}
	}
}

func TestSampleLeadsCoverEveryBandAndMissingFields(t *testing.T) {
	ranked, err := Rank(SampleLeads())
	if err != nil {
		t.Fatalf("Rank(SampleLeads()) error = %v", err)
	}
	bands := map[string]bool{}
	missing := map[string]bool{}
	for _, score := range ranked {
		bands[score.Band] = true
		for _, field := range score.Missing {
			missing[field] = true
		}
	}
	for _, band := range []string{"high", "medium", "low"} {
		if !bands[band] {
			t.Errorf("SampleLeads() has no %q band", band)
		}
	}
	for _, field := range []string{"timeline_days", "interactions"} {
		if !missing[field] {
			t.Errorf("SampleLeads() has no missing %q field", field)
		}
	}
}

func cloneLeadsForTest(leads []Lead) []Lead {
	cloned := append([]Lead(nil), leads...)
	for i := range cloned {
		if cloned[i].TimelineDays != nil {
			value := *cloned[i].TimelineDays
			cloned[i].TimelineDays = &value
		}
		if cloned[i].Interactions != nil {
			value := *cloned[i].Interactions
			cloned[i].Interactions = &value
		}
	}
	return cloned
}
