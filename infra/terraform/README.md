# Terraform Placeholder

This folder is intentionally a placeholder structure. It gives agents a clear place to add Terraform when an instantiated project chooses Terraform, but the scaffold does not ship deployable Terraform code.

Do not add `.tf` files until the target project has selected:

- cloud or hosting provider
- remote state backend and locking strategy
- environment names and ownership
- stack boundaries
- secret management strategy
- validation, plan, apply, and rollback workflow

## Structure

```text
infra/terraform
├── environments/   # environment roots such as dev, staging, prod
├── stacks/         # deployable units such as platform, backend, web, observability
└── modules/        # reusable Terraform modules
```

## Instantiation Rules

When Terraform is selected:

1. Add exact provider versions and required Terraform version constraints.
2. Configure remote state before the first shared plan.
3. Keep secrets out of source control.
4. Keep environment-specific values in variables or documented defaults.
5. Add `terraform fmt`, `terraform validate`, and plan review to CI.
6. Document exact commands in `docs/deployment.md`.
