# Deployment Contract

The scaffold is provider-neutral. It should make local development, tests, runtime configuration, and observability clear, but it should not add deployable infrastructure code until a real project selects its runtime platform.

## What Must Exist After Instantiation

Every selected service, worker, app, and operational tool must document:

- runtime command
- build command
- required environment variables
- health check or process liveness check
- readiness check for dependencies
- smoke test command
- metrics, logs, and trace wiring
- rollback or recovery path for production changes

## Backend And Web Connection

Backend services must expose documented API boundaries and typed clients where the selected stack supports them. Web and mobile apps should read API locations from environment variables such as `PUBLIC_API_BASE_URL`; do not bake hostnames or ports into source code.

For local development, `make infra-up` starts dependencies and observability. Project-specific services and web apps should then be runnable with their own documented commands, and smoke tests should verify that the selected web surface can reach the selected backend boundary.

The scaffold's first runnable path is:

```sh
docker compose up --build
```

It serves the web app on `http://localhost:3000`, the tasks API on
`http://localhost:8080`, and a background worker that processes jobs written by
task creation.

## Adding Runtime Manifests

When a project chooses its hosting platform:

1. Add only the manifests and commands required for that platform.
2. Keep runtime configuration typed and documented.
3. Keep secrets out of source control.
4. Add health, readiness, and metrics wiring to each deployable.
5. Add a smoke test for the deployed backend and web path.
6. Update CI so deployable artifacts are built and checked before merge.

Do not keep unused deployment adapters in the scaffolded project.

## Terraform Placeholder

`infra/terraform` is a tracked placeholder structure. It exists so an LLM agent has an obvious place to add Terraform later, but it intentionally contains no `.tf` implementation files.

Use it only after the project chooses Terraform. Before writing Terraform code, identify:

- target provider or providers
- remote state backend and locking strategy
- environments that should exist
- stacks to deploy: platform, backend, web, observability, or project-specific equivalents
- reusable modules required for network, compute, data, security, and observability
- secret management strategy
- exact validation, plan, apply, and destroy commands

When Terraform is instantiated, add real `.tf` files under the appropriate `environments`, `stacks`, and `modules` folders, document commands here, and add CI checks for `terraform fmt`, `terraform validate`, and safe plan review.
