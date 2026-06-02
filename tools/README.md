# Tools

Tools are organized by developer or operational capability, not by implementation language.

Use folders like:

- `developer-experience`
- `migrations`
- `release`
- `data-repair`
- `codegen`
- `observability`

When a project is instantiated, keep the tool capabilities that matter and remove the rest. Pick implementation languages inside the retained tool domains only after the workflow is clear.

Required tool test categories are tracked under `tools/tests/` and described in `docs/testing-matrix.md`.
