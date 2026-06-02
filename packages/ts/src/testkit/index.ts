export function assertEqual<T>(got: T, want: T): void {
  if (got !== want) {
    throw new Error(`got ${String(got)}, want ${String(want)}`);
  }
}
