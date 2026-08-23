import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "../../../app/AppProviders";
import { issueSearchFixture } from "../../../test/issue-fixtures";
import {
  createDefaultSearchFilters,
  type DecodedSearchLocation,
} from "../model/search-filters";
import { useIssueSearch } from "./useIssueSearch";

afterEach(() => {
  vi.unstubAllGlobals();
});

function location(
  overrides: Partial<DecodedSearchLocation> = {},
): DecodedSearchLocation {
  return {
    errors: {},
    filters: createDefaultSearchFilters("octocat"),
    shouldSearch: true,
    valid: true,
    ...overrides,
  };
}

describe("useIssueSearch", () => {
  it("posts the typed request and forwards query cancellation", async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(issueSearchFixture), {
        headers: { "Content-Type": "application/json" },
        status: 200,
      }),
    );
    vi.stubGlobal("fetch", request);

    const { result } = renderHook(() => useIssueSearch(location()), {
      wrapper: AppProviders,
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    const [url, options] = request.mock.calls[0] ?? [];
    expect(url).toBe("/api/issues/search?page=1&perPage=20");
    expect(options?.signal).toBeInstanceOf(AbortSignal);
    expect(options?.body).toEqual(expect.any(String));
    expect(JSON.parse(options?.body as string)).toMatchObject({
      maximumDifficulty: 3,
      minimumStars: 10,
      username: "octocat",
    });
    expect(request).toHaveBeenCalledTimes(1);
  });

  it("retries once with relaxed filters after an empty exact search", async () => {
    const empty = structuredClone(issueSearchFixture);
    empty.data.items = [];
    empty.data.pagination = {
      ...empty.data.pagination,
      hasNext: false,
      total: 0,
      totalPages: 0,
    };
    const request = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(
        new Response(JSON.stringify(empty), {
          headers: { "Content-Type": "application/json" },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(issueSearchFixture), {
          headers: { "Content-Type": "application/json" },
        }),
      );
    vi.stubGlobal("fetch", request);

    const { result } = renderHook(() => useIssueSearch(location()), {
      wrapper: AppProviders,
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[1]).toBe(true);
    expect(request).toHaveBeenCalledTimes(2);
    const fallbackBody = JSON.parse(
      request.mock.calls[1]?.[1]?.body as string,
    ) as Record<string, unknown>;
    expect(fallbackBody).toMatchObject({
      frameworks: [],
      includeStale: true,
      languages: [],
      maximumDifficulty: 5,
      minimumStars: 0,
      updatedWithinDays: 3650,
    });
  });

  it("does not request data before search or for invalid URL state", () => {
    const request = vi.fn<typeof fetch>();
    vi.stubGlobal("fetch", request);

    renderHook(
      () =>
        useIssueSearch(
          location({
            shouldSearch: false,
          }),
        ),
      { wrapper: AppProviders },
    );
    renderHook(
      () =>
        useIssueSearch(
          location({
            errors: { form: "Invalid shared URL" },
            valid: false,
          }),
        ),
      { wrapper: AppProviders },
    );

    expect(request).not.toHaveBeenCalled();
  });
});
