#!/usr/bin/env bash
set -euo pipefail

required=(git gh node pnpm go cargo java docker)

for tool in "${required[@]}"; do
  if command -v "$tool" >/dev/null 2>&1; then
    printf "ok   %s\n" "$tool"
  else
    printf "miss %s\n" "$tool"
  fi
done

printf "\nRun 'pnpm install' before TypeScript commands and 'make infra-up' before integration work.\n"
