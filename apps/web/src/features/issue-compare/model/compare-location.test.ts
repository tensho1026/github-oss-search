import { describe, expect, it } from "vitest";

import {
  decodeCompareLocation,
  encodeCompareLocation,
  maximumComparedIssues,
} from "./compare-location";

describe("compare location", () => {
  const references = [
    { issueNumber: 1, owner: "acme", repository: "rocket" },
    { issueNumber: 2, owner: "openai", repository: "sdk" },
    { issueNumber: 3, owner: "octo", repository: "tool" },
  ];

  it("round-trips a bounded shareable comparison", () => {
    const encoded = encodeCompareLocation(
      [
        ...references,
        { issueNumber: 4, owner: "extra", repository: "ignored" },
      ],
      [" Go ", "", "React"],
      "/search?username=octocat",
    );
    expect(encoded.getAll("issue")).toHaveLength(maximumComparedIssues);
    expect(encoded.getAll("skills")).toEqual(["Go", "React"]);
    const decoded = decodeCompareLocation(encoded);
    expect(decoded).toMatchObject({
      references,
      returnTo: "/search?username=octocat",
      skills: ["Go", "React"],
      valid: true,
    });
  });

  it.each([
    [new URLSearchParams(), "Select two or three"],
    [
      new URLSearchParams(
        "issue=a%2Fb%231&issue=c%2Fd%232&issue=e%2Ff%233&issue=g%2Fh%234",
      ),
      "Select two or three",
    ],
    [new URLSearchParams("issue=bad&issue=c%2Fd%232"), "invalid"],
    [
      new URLSearchParams(
        "issue=a%2Fb%239999999999999999999999&issue=c%2Fd%232",
      ),
      "invalid",
    ],
    [new URLSearchParams("issue=A%2FB%231&issue=a%2Fb%231"), "unique"],
  ])("rejects malformed references", (parameters, message) => {
    const decoded = decodeCompareLocation(parameters);
    expect(decoded.valid).toBe(false);
    expect(decoded.message).toContain(message);
  });

  it("rejects excessive or oversized skill evidence", () => {
    const excessive = encodeCompareLocation(
      references.slice(0, 2),
      [],
      "/search",
    );
    for (let index = 0; index < 21; index += 1)
      excessive.append("skills", `skill-${index}`);
    expect(decodeCompareLocation(excessive).valid).toBe(false);

    const oversized = encodeCompareLocation(
      references.slice(0, 2),
      [],
      "/search",
    );
    oversized.append("skills", "x".repeat(65));
    expect(decodeCompareLocation(oversized).valid).toBe(false);
  });

  it("falls back from unsafe return paths and omits them while encoding", () => {
    const encoded = encodeCompareLocation(
      references.slice(0, 2),
      [],
      "https://evil.test",
    );
    expect(encoded.has("returnTo")).toBe(false);
    encoded.set("returnTo", "//evil.test");
    expect(decodeCompareLocation(encoded).returnTo).toBe("/search");
    encoded.set("returnTo", `/${"x".repeat(2048)}`);
    expect(decodeCompareLocation(encoded).returnTo).toBe("/search");
  });
});
