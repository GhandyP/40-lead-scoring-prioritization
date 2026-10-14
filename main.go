package main

import (
	"context"
	"log"
	"os"

	"github.com/GhandyP/adk-go-enterprise-cookbook/internal/adkhelper"
)

func main() {
	if err := adkhelper.Run(
		context.Background(),
		os.Args[1:],
		"case_40_lead_scoring_prioritization",
		"A text-only starting point for lead scoring and prioritization.",
		"Explain a lead priority from the features supplied by the user. Keep score reasoning tied to those features, flag missing data, and do not claim CRM or predictive-model access.",
	); err != nil {
		log.Fatal(err)
	}
}
