# Lead priority workbench

A small Go portfolio project that makes lead prioritization explainable: a ranked table of fictional prospects, visible scoring reasons, and explicit missing-data warnings. Runs locally with **no API keys, CRM, database, JavaScript build tools, or external Go dependencies**.

## Run the demo

Install [Go 1.22 or later](https://go.dev/doc/install), then:

```bash
git clone https://github.com/GhandyP/40-lead-scoring-prioritization.git
cd 40-lead-scoring-prioritization
go run .
```

Open **http://127.0.0.1:8080**. Stop with `Ctrl+C`. The server listens only on the loopback interface; this is a local demo, not a production deployment.

- Search by name, company, or lead ID, and filter with `?band=high`, `?band=medium`, or `?band=low`.
- Omit `band` to show all leads; an unknown non-empty value returns HTTP 400 with valid choices.
- Read the evidence behind each score and see which inputs are missing.
- Choose **Clear filters** to restore the full fictional dataset.
- Open [`/api/leads`](http://127.0.0.1:8080/api/leads) for the complete ranked JSON dataset. Dashboard filters do not change the API results.

## Scoring rules

Rules are versioned as `v1`. Scores are illustrative points, **not conversion probabilities**.

| Signal | Points |
| --- | ---: |
| Demo requested | 30 |
| Budget confirmed | 25 |
| Timeline 0–30 days | 25 |
| Timeline 31–90 days | 10 |
| Timeline over 90 days | 0 |
| Recorded interactions | 5 each, capped at 4 (20 points) |

High priority is **70–100**, medium **40–69**, and low **0–39**. Results sort by score descending, then lead ID ascending. The sample's expected order is Maya Chen (100), Sam Taylor (55), Elliot Park (50), Jordan Rivera (15).

Missing timeline or interaction values contribute no points but are explicitly labeled missing, rather than represented as known zeros. Zero is valid: a timeline of zero means immediate interest; zero interactions means none recorded. Negative numbers, blank IDs and duplicate IDs are rejected by `Rank` without changing its input. Each response includes reasons, missing fields, and the scoring version.

## Architecture

```text
Browser GET /?q=...&band=... ──> NewHandler ──> embedded HTML + CSS
Client  GET /api/leads      ──> NewHandler ──> JSON ranking
                                   │
                           Rank(SampleLeads())
                                   │
                      validated, deterministic rules
```

| File | Responsibility |
| --- | --- |
| `scorer.go` | Domain validation, rule explanations, priority bands and stable ranking |
| `fixtures.go` | Fictional sample leads; no external records |
| `web.go`, `web/` | GET-only HTTP routes, filters and embedded presentation |
| `main.go` | Local server entry point |
| `scorer_test.go`, `web_test.go` | Domain and HTTP regression checks |

HTML and CSS are embedded in the binary, so serving them does not depend on the working directory. The browser requires no JavaScript. Unknown routes return 404; non-GET methods on known routes return 405 with `Allow: GET`.

## Verify

```bash
go test ./...
go vet ./...
go build -o bin/lead-workbench .
./bin/lead-workbench
```

Tests cover score boundaries, missing versus zero inputs, duplicate and invalid records, deterministic ties, input immutability, valid filters, invalid-band rejection, empty results, escaped search input, JSON output, embedded CSS and HTTP methods. Chromium desktop (1440×1100) and mobile (390×844) checks confirmed the ranked data, combined filters, empty and reset states, loaded CSS, and score explanations fitting on mobile without horizontal scrolling.

## Scope and limitations

- All names and companies are fictional. No customer data is loaded, stored, assigned, contacted or sent to an external service.
- Rules only use demo requests, confirmed budget, timeline and interactions. Names and companies are display/search fields, not scoring inputs. This is not proof of fairness or a calibrated model.
- Scores are advisory; missing signals can lower a lead's rank. Do not use the demo to make real sales decisions without validating your data, business rules and bias risks.
- No authentication, persistent storage, CRM adapter, AI model or production hardening is included.

This replaces a non-standalone Gemini cookbook stub with a self-contained example. See the [ODD plan and verification log](odd/tasks/portfolio-demo.md) for the implementation record.
