import { describe, expect, it } from "vitest";

import { issueDetailFixture } from "../../../test/issue-fixtures";
import {
  compareHighlightWinners,
  issueHighlightKey,
} from "./compare-highlights";

describe("compare highlights", () => {
  it("marks a unique relative better value without ranking ties", () => {
    const left = structuredClone(issueDetailFixture.data);
    const right = structuredClone(issueDetailFixture.data);
    right.repository.name = "other-service";
    right.issue.number = 7;
    right.recommendation.skillMatch.percentage = 40;
    right.recommendation.maintainerResponse.firstIssueResponse.medianSeconds = 80_000;
    right.recommendation.stale.state = "stale";

    expect(compareHighlightWinners([left, right])).toEqual({
      response: issueHighlightKey(left),
      skillMatch: issueHighlightKey(left),
      stale: issueHighlightKey(left),
    });
    right.recommendation.skillMatch.percentage =
      left.recommendation.skillMatch.percentage;
    expect(compareHighlightWinners([left, right]).skillMatch).toBeUndefined();
  });
});
