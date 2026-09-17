import { describe, expect, it } from "vitest";

import {
  appendSearchSelection,
  decodeSearchSelection,
  searchSelectionParameterName,
  toggleSearchSelection,
} from "./search-selection";

const first = { issueNumber: 1, owner: "acme", repository: "rocket" };
const second = { issueNumber: 2, owner: "octo", repository: "tool" };
const third = { issueNumber: 3, owner: "openai", repository: "sdk" };
const fourth = { issueNumber: 4, owner: "extra", repository: "ignored" };

describe("search selection", () => {
  it("round-trips at most three unique issue references", () => {
    const parameters = new URLSearchParams("language=Go");
    appendSearchSelection(parameters, [first, first, second, third, fourth]);
    expect(parameters.getAll(searchSelectionParameterName)).toEqual([
      "acme/rocket#1",
      "octo/tool#2",
      "openai/sdk#3",
    ]);
    expect(parameters.get("language")).toBe("Go");
    expect(decodeSearchSelection(parameters)).toEqual([first, second, third]);
  });

  it("ignores malformed references instead of failing the search URL", () => {
    const parameters = new URLSearchParams(
      `${searchSelectionParameterName}=bad&${searchSelectionParameterName}=acme/rocket%231`,
    );
    expect(decodeSearchSelection(parameters)).toEqual([first]);
  });

  it("toggles selection without exceeding the compare bound", () => {
    expect(toggleSearchSelection([], first)).toEqual([first]);
    expect(toggleSearchSelection([first], first)).toEqual([]);
    expect(toggleSearchSelection([first, second, third], fourth)).toEqual([
      first,
      second,
      third,
    ]);
  });
});
