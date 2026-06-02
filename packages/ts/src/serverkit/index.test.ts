import { describe, expect, it } from "vitest";
import { healthPayload } from "./index";

describe("serverkit", () => {
  it("creates a normalized health payload", () => {
    expect(healthPayload("project_scaffold")).toEqual({
      service: "Project Scaffold",
      status: "ok",
    });
  });
});
