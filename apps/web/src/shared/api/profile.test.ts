import { afterEach, describe, expect, it, vi } from "vitest";

import type { ApiClient } from "./client";
import {
  getGitHubUser,
  getProfileAnalysis,
  getProfileSnapshot,
} from "./profile";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("profile API adapters", () => {
  it("maps each profile request to its canonical endpoint", async () => {
    const get = vi.fn().mockResolvedValue({});
    const client = { get } as unknown as ApiClient;

    await getGitHubUser("octo cat", undefined, client);
    await getProfileAnalysis("octo cat", undefined, client);
    await getProfileSnapshot("octo cat", undefined, client);

    expect(get.mock.calls.map(([path]) => path)).toEqual([
      "/api/github/users/octo%20cat",
      "/api/github/users/octo%20cat/profile-analysis",
      "/api/github/users/octo%20cat/profile-snapshot",
    ]);
  });

  it("supports the default client for profile requests", async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("{}", {
        headers: { "Content-Type": "application/json" },
        status: 200,
      }),
    );
    vi.stubGlobal("fetch", request);

    await getGitHubUser("octo cat");

    expect(request).toHaveBeenCalledWith(
      "/api/github/users/octo%20cat",
      expect.objectContaining({ method: "GET" }),
    );
  });
});
