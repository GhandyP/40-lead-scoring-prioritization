package main

import (
	"fmt"
	"sort"
	"strings"
)

const scoreVersionV1 = "v1"

// Lead contains the fictional lead attributes used by the advisory scoring rules.
type Lead struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Company         string `json:"company"`
	RequestedDemo   bool   `json:"requested_demo"`
	BudgetConfirmed bool   `json:"budget_confirmed"`
	TimelineDays    *int   `json:"timeline_days"`
	Interactions    *int   `json:"interactions"`
}

// Score describes a lead's illustrative score, band, and the data behind it.
type Score struct {
	Lead    Lead     `json:"lead"`
	Score   int      `json:"score"`
	Band    string   `json:"band"`
	Reasons []string `json:"reasons"`
	Missing []string `json:"missing"`
	Version string   `json:"version"`
}

// Rank validates and ranks leads using the illustrative v1 rules. It does not
// modify the input; results are ordered by descending score and then ascending ID.
func Rank(leads []Lead) ([]Score, error) {
	seenIDs := make(map[string]struct{}, len(leads))
	for i, lead := range leads {
		if strings.TrimSpace(lead.ID) == "" {
			return nil, fmt.Errorf("lead at index %d has a blank ID", i)
		}
		if _, exists := seenIDs[lead.ID]; exists {
			return nil, fmt.Errorf("duplicate lead ID %q", lead.ID)
		}
		seenIDs[lead.ID] = struct{}{}
		if lead.TimelineDays != nil && *lead.TimelineDays < 0 {
			return nil, fmt.Errorf("lead %q has a negative timeline", lead.ID)
		}
		if lead.Interactions != nil && *lead.Interactions < 0 {
			return nil, fmt.Errorf("lead %q has a negative interaction count", lead.ID)
		}
	}

	ranked := make([]Score, 0, len(leads))
	for _, lead := range leads {
		score := Score{
			Lead:    copyLeadValues(lead),
			Reasons: make([]string, 0, 4),
			Missing: make([]string, 0, 2),
			Version: scoreVersionV1,
		}

		if lead.RequestedDemo {
			score.Score += 30
			score.Reasons = append(score.Reasons, "Demo requested (+30)")
		} else {
			score.Reasons = append(score.Reasons, "Demo not requested (+0)")
		}
		if lead.BudgetConfirmed {
			score.Score += 25
			score.Reasons = append(score.Reasons, "Budget confirmed (+25)")
		} else {
			score.Reasons = append(score.Reasons, "Budget not confirmed (+0)")
		}

		switch {
		case lead.TimelineDays == nil:
			score.Missing = append(score.Missing, "timeline_days")
			score.Reasons = append(score.Reasons, "Timeline unavailable (missing)")
		case *lead.TimelineDays <= 30:
			score.Score += 25
			score.Reasons = append(score.Reasons, "Timeline within 30 days (+25)")
		case *lead.TimelineDays <= 90:
			score.Score += 10
			score.Reasons = append(score.Reasons, "Timeline 31–90 days (+10)")
		default:
			score.Reasons = append(score.Reasons, "Timeline over 90 days (+0)")
		}

		switch {
		case lead.Interactions == nil:
			score.Missing = append(score.Missing, "interactions")
			score.Reasons = append(score.Reasons, "Interactions unavailable (missing)")
		case *lead.Interactions >= 4:
			score.Score += 20
			score.Reasons = append(score.Reasons, "Interactions: 4 or more (+20; capped at 4)")
		default:
			points := *lead.Interactions * 5
			score.Score += points
			score.Reasons = append(score.Reasons, fmt.Sprintf("Interactions: %d (+%d)", *lead.Interactions, points))
		}

		switch {
		case score.Score >= 70:
			score.Band = "high"
		case score.Score >= 40:
			score.Band = "medium"
		default:
			score.Band = "low"
		}
		ranked = append(ranked, score)
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return ranked[i].Lead.ID < ranked[j].Lead.ID
	})

	return ranked, nil
}

func copyLeadValues(lead Lead) Lead {
	copied := lead
	if lead.TimelineDays != nil {
		timelineDays := *lead.TimelineDays
		copied.TimelineDays = &timelineDays
	}
	if lead.Interactions != nil {
		interactions := *lead.Interactions
		copied.Interactions = &interactions
	}
	return copied
}
