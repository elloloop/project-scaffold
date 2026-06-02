# Google3/Piper-Style Layout

This repository follows the part of the google3/Piper model that matters for a reusable scaffold:

- one repository tree
- product/domain/capability paths before language or framework choices
- local ownership by directory
- tests close to the area they protect
- framework options represented explicitly but not preselected
- language-specific folders only for shared packages
- agent instructions versioned with the code
- build-system-neutral package boundaries

Example:

```text
services/
  identity/
  billing/
  platform-gateway/
workers/
  email-delivery/
  search-indexing/
tools/
  migrations/
  release/
packages/
  go/
  ts/
  rust/
  java/
```

New projects should remove unused domains, capabilities, framework options, and shared language lanes during setup.

This scaffold includes no repository-wide build metadata. New projects should add only the build system required by their selected stack.
