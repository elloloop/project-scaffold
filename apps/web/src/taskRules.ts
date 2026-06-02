export function normalizeTaskTitle(value: string): string {
  return value.trim().replace(/\s+/g, " ");
}

export function canSubmitTaskTitle(value: string): boolean {
  const normalized = normalizeTaskTitle(value);
  return normalized.length > 0 && normalized.length <= 120;
}
