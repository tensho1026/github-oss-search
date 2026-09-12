import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "../api/client";
import { shouldRetryQuery } from "./retry-policy";

describe("shouldRetryQuery", () => {
  it.each([400, 401, 403, 404, 429])("does not retry status %d", (status) => {
    expect(
      shouldRetryQuery(
        0,
        new ApiError({ code: "TEST", message: "test", status }),
      ),
    ).toBe(false);
  });

  it("retries one transient failure only", () => {
    expect(shouldRetryQuery(0, new Error("network"))).toBe(true);
    expect(shouldRetryQuery(1, new Error("network"))).toBe(false);
  });

  it.each([
    [401, 1],
    [429, 1],
    [500, 2],
  ])(
    "issues the expected request count for status %d",
    async (status, count) => {
      const request = vi
        .fn()
        .mockRejectedValue(
          new ApiError({ code: "TEST", message: "test", status }),
        );
      const client = new QueryClient({
        defaultOptions: {
          queries: { retry: shouldRetryQuery, retryDelay: 0 },
        },
      });

      await expect(
        client.fetchQuery({ queryFn: request, queryKey: ["retry", status] }),
      ).rejects.toBeInstanceOf(ApiError);
      expect(request).toHaveBeenCalledTimes(count);
      client.clear();
    },
  );
});
