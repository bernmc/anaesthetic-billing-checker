# Acceptance Tests

Use these cases to verify behavior after build.

## Case A: Valve replacement style case with missing arterial insertion
Input:
20560
25005
22012
22012
22012
22015
22020
22020
22052
22054
25014
23360
17620

Expected:
- Should identify likely valve-related pattern (if matching logic implemented).
- Should raise QUERY about monitoring/insertion consistency if 22025 absent.
- Should not fabricate unknown values.
- Should produce schedule total from source data where available.
- Should produce unit-based totals only from known unit mappings.

## Case B: TURP elderly case
Input:
20914
25014
17610
25000
23045

Expected:
- Should parse all codes.
- Should include MBS descriptions from current data source.
- Should evaluate safety checks:
  - one pre-op (pass),
  - procedure present (pass),
  - one time code (pass, if 23045 treated as time code),
  - any missing mappings should be transparent.

## Case C: Double pre-op hard fail
Input:
17610
17620
20914
23045

Expected:
- HARD FAIL on pre-op rule (more than one pre-op).

## Case D: Missing time hard fail
Input:
17610
20914
25014

Expected:
- HARD FAIL on time code missing.

## Case E: High audit-risk code trigger
Input:
17610
20560
22014
23045

Expected:
- High audit-risk warning present and clearly visible.

## Case F: Data freshness
Action:
- Trigger in-UI update button.

Expected:
- Update succeeds.
- Metadata updates (source + generated timestamp + item count).
- App continues functioning after refresh.

## Pass criteria
- No hallucinated code descriptions.
- No hidden assumptions for unknown fields.
- Traffic-light panel sorted by severity.
- Non-coder can run and stop app without memorizing complex commands.
