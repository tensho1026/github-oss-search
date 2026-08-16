# OSS health dashboard methodology

The issue detail page reports four independent repository indicators. They are
diagnostic heuristics, not a universal ranking, endorsement, or safety
certification. Score rules are pinned to version `2026-08-01`.

## Category calculation

Each component has a fixed percentage weight and a normalized 0–100 score. A
category is the weighted mean of available components. When some components
are unavailable, the number is labeled `partial` and the denominator contains
only known weights. No known components produces `unavailable` and a null
score.

| Category          | Components and weights                                                                       |
| ----------------- | -------------------------------------------------------------------------------------------- |
| Activity          | update recency 50%, default-branch CI 25%, sampled PR merge ratio 25%                        |
| Community         | sampled contributors 35%, issue response 25%, PR review 25%, contributing guide 15%          |
| Beginner friendly | contributing 30%, README 20%, conduct 15%, tests 15%, CI 10%, issue response 10%             |
| Security          | OpenSSF Maintained 25%, Code-Review 25%, Contributors 15%, CI-Tests 20%, Vulnerabilities 15% |

Complete inputs produce high confidence, partial inputs medium confidence, and
no inputs low confidence. Every category returns its components, weights,
source, analysis time, and warnings so a partial score cannot silently look
complete.

Issue search cards show the same four category summaries from GitHub evidence
without starting third-party requests. Security therefore remains unavailable
on a card and is loaded only when the user opens repository detail.

## OpenSSF Scorecard boundary

The API requests the official published Scorecard REST result for
`github.com/{owner}/{repository}`. It accepts at most 50 checks and a 1 MiB
response, normalizes only check name and score, maps upstream `-1` to
unavailable, and rejects unexpected score ranges or malformed schemas. React
never receives the upstream payload, reasons, or arbitrary documentation URLs.

Successful and absent results are cached for six hours with a 500-entry bound.
Requests time out after three seconds. Only 502 and 503 receive one immediate
retry; rate limits and malformed responses do not retry. A third-party failure
returns an unavailable Security category while GitHub-based categories remain
usable. Results older than 30 days are labeled stale, and the upstream
Scorecard version is preserved when present.

OpenSSF Scorecard itself is heuristic. A high Security indicator does not
guarantee that a repository or its dependencies are safe. Vulnerability
scanning and dependency remediation are outside this feature.
