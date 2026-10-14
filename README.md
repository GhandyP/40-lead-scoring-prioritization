# 40 — Lead scoring and prioritization

## Goal

Explain lead priority from supplied features without claiming CRM access or a calibrated predictive scoring model.

## Run the current prototype

From this case directory:

```bash
export GOOGLE_API_KEY="your-key"
go run .
```

`internal/adkhelper.Run` uses ADK Go v2.2.0, Gemini, and `gemini-flash-latest`; set `ADK_MODEL` to override the model.

Sample prompt:

```text
Prioritize these anonymized leads and explain the evidence for each.
```

## What exists now

- `main.go (exists)` sends anonymized lead text to one focused Gemini ADK agent.
- The response has no lead ID, feature schema, score calibration, reason codes, or deterministic tie-breaking.
- This is intentionally text-only: it does not load CRM records, train a scoring model, or change lead ownership/status.

## Missing elements

- **Contract:** Define `LeadBatch{leads[{id,features}], scoringVersion}` -> `LeadScore{id, score, band, reasons[], confidence}`.
- **ADK layer:** Add a schema-constrained explanation agent and batch session; use a read-only `get_lead_features` tool, never an unapproved assignment tool.
- **Domain boundary:** Add CRM feature adapter, feature normalizer, calibrated rules/model scorer, and ownership-update adapter.
- **Fixtures and validation:** Cover missing features, class imbalance, boundary scores, duplicate leads, and protected attributes; test calibration and exclusion rules.
- **Safety:** Keep scores advisory, audit feature provenance, exclude protected attributes, and require sales-owner approval before routing or contacting.

## Target structure

```text
40-lead-scoring-prioritization/
├── main.go (exists)
├── README.md (exists)
├── schema.go (next)
├── scorer.go (next)
├── adapters/crm.go (next)
├── testdata/leads.json (next)
└── main_test.go (next)
```

## Implementation order

1. Define lead features, score bands, reason codes, and scoring version.
2. Implement a deterministic baseline and protected-feature filter.
3. Add read-only CRM retrieval and explanation session.
4. Test missing data, duplicate leads, thresholds, and calibration fixtures.
5. Add dry-run routing only after bias and owner-approval review.

## Done when

- [ ] Every score has feature provenance, version, band, and reason codes.
- [ ] Missing or protected features cannot silently change a decision.
- [ ] Fixtures report deterministic thresholds and calibration behavior.
- [ ] CRM routing or contact actions require a human sales owner.

## Next improvement

Build the versioned deterministic scorer and protected-feature tests before connecting the CRM.
