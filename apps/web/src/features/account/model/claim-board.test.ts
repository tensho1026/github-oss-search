import { describe, expect, it } from "vitest";

import type { IssueClaim } from "../../../shared/api/generated";
import { claimMoveRequest } from "./claim-board";

const claim = {
  archived: false,
  id: "123e4567-e89b-42d3-a456-426614174000",
  issueNumber: 42,
  pullRequest: null,
  repositoryName: "typed-service",
  repositoryOwner: "octocat",
  status: "researching",
  version: 3,
} as IssueClaim;

describe("claim Kanban moves", () => {
  it("builds an optimistic-versioned personal workflow update", () => {
    expect(claimMoveRequest(claim, "implementing")).toEqual({
      ok: true,
      request: {
        archived: false,
        pullRequest: null,
        status: "implementing",
        version: 3,
      },
    });
  });

  it("requires a linked pull request for submitted and merged columns", () => {
    expect(claimMoveRequest(claim, "pr_submitted")).toEqual({
      ok: false,
      reason: "pull_request_required",
    });
    expect(claimMoveRequest(claim, "merged")).toEqual({
      ok: false,
      reason: "pull_request_required",
    });
  });

  it("does not send a mutation when dropped into its current column", () => {
    expect(claimMoveRequest(claim, "researching")).toEqual({
      ok: false,
      reason: "unchanged",
    });
  });
});
