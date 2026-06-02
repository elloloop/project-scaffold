import { describe, expect, it } from "vitest";

import { canSubmitTaskTitle, normalizeTaskTitle } from "./taskRules";

describe("task title rules", () => {
  it("normalizes whitespace", () => {
    expect(normalizeTaskTitle("  Ship   scaffold  ")).toBe("Ship scaffold");
  });

  it("rejects blank and overly long titles", () => {
    expect(canSubmitTaskTitle("   ")).toBe(false);
    expect(canSubmitTaskTitle("x".repeat(121))).toBe(false);
    expect(canSubmitTaskTitle("Write tests")).toBe(true);
  });
});
