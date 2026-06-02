# New Project Checklist

Use this after cloning the scaffold.

1. Rename the repository and root package identifiers.
2. Update `README.md` with the product purpose and first-run flow.
3. Update `configs/env/.env.example` with required runtime configuration.
4. Choose web/mobile/desktop frameworks only when the project needs those surfaces.
5. Delete unused option folders and unused shared package language lanes.
6. Keep every applicable test category from `docs/testing-matrix.md`.
7. Keep at least one test command per retained language lane.
8. Replace sample service names with product-specific names.
9. Wire retained services, workers, and apps into `docs/observability.md`.
10. Choose deployment tooling only after the runtime platform is selected.
11. If Terraform is chosen, populate `infra/terraform` with real providers, state, modules, environment roots, and CI validation.
12. Update `AGENTS.md` with project-specific rules.
13. Create real review teams and convert `.github/CODEOWNERS.example` into `.github/CODEOWNERS`.
14. Enable branch protection with required PRs, code-owner review, and CI checks.
15. Add ADRs for major structure, runtime, or persistence decisions.
16. Confirm `make test` passes before the first real feature branch.
