import { describe, expect, it } from "vitest";

import {
  appendSavedSearchId,
  decodeSavedSearchId,
  savedSearchParameterName,
} from "./saved-search-location";

const validId = "00000000-0000-4000-8000-000000000001";

describe("saved search location", () => {
  it("round-trips a single UUID", () => {
    const parameters = new URLSearchParams();
    appendSavedSearchId(parameters, validId);
    expect(parameters.get(savedSearchParameterName)).toBe(validId);
    expect(decodeSavedSearchId(parameters)).toBe(validId);
  });

  it("ignores missing, duplicated, or malformed identifiers", () => {
    expect(decodeSavedSearchId(new URLSearchParams())).toBeUndefined();
    const duplicated = new URLSearchParams();
    duplicated.append(savedSearchParameterName, validId);
    duplicated.append(savedSearchParameterName, validId);
    expect(decodeSavedSearchId(duplicated)).toBeUndefined();
    expect(
      decodeSavedSearchId(
        new URLSearchParams(`${savedSearchParameterName}=not-a-uuid`),
      ),
    ).toBeUndefined();
  });
});
