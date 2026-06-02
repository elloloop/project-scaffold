# Workers

Workers are organized by background capability or product workflow, not by implementation language.

Use folders like:

- `email-delivery`
- `data-ingestion`
- `billing-reconciliation`
- `search-indexing`
- `scheduled-jobs`
- `ai-processing`

When a project is instantiated, keep the worker capabilities that matter and remove the rest. Pick implementation languages inside the retained worker domains only after the product need is clear.

Required worker test categories are tracked under `workers/tests/` and described in `docs/testing-matrix.md`.
