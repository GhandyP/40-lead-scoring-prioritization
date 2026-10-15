package main

// SampleLeads returns fictional portfolio-demo data with complete and missing
// optional values. Scores are illustrative rules, not predictions.
func SampleLeads() []Lead {
	return []Lead{
		{
			ID:              "lead-101",
			Name:            "Maya Chen",
			Company:         "Northstar Labs",
			RequestedDemo:   true,
			BudgetConfirmed: true,
			TimelineDays:    sampleIntPointer(14),
			Interactions:    sampleIntPointer(5),
		},
		{
			ID:              "lead-102",
			Name:            "Elliot Park",
			Company:         "Harborlight Analytics",
			RequestedDemo:   true,
			BudgetConfirmed: false,
			TimelineDays:    sampleIntPointer(45),
			Interactions:    sampleIntPointer(2),
		},
		{
			ID:              "lead-103",
			Name:            "Jordan Rivera",
			Company:         "Cedarline Systems",
			RequestedDemo:   false,
			BudgetConfirmed: false,
			TimelineDays:    sampleIntPointer(90),
			Interactions:    sampleIntPointer(1),
		},
		{
			ID:              "lead-104",
			Name:            "Sam Taylor",
			Company:         "Bluewood Studio",
			RequestedDemo:   true,
			BudgetConfirmed: true,
		},
	}
}

func sampleIntPointer(value int) *int {
	return &value
}
