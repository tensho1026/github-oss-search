import type { IssueDetail } from "../../../shared/api/generated";

export type CompareHighlightField = "response" | "skillMatch" | "stale";

const staleRank: Record<string, number> = {
  aging: 1,
  fresh: 0,
  stale: 3,
  unknown: 2,
};

export function compareHighlightWinners(
  issues: readonly IssueDetail[],
): Partial<Record<CompareHighlightField, string>> {
  if (issues.length < 2) {
    return {};
  }
  return {
    ...uniqueWinner(issues, "skillMatch", (issue) => {
      return issue.recommendation.skillMatch.percentage;
    }),
    ...uniqueWinner(issues, "response", (issue) => {
      const response = issue.recommendation.maintainerResponse;
      if (
        response.status !== "available" ||
        response.firstIssueResponse.medianSeconds === null
      ) {
        return;
      }
      return -response.firstIssueResponse.medianSeconds;
    }),
    ...uniqueWinner(issues, "stale", (issue) => {
      const rank = staleRank[issue.recommendation.stale.state];
      return rank === undefined ? undefined : -rank;
    }),
  };
}

export function issueHighlightKey(issue: IssueDetail): string {
  return `${issue.repository.owner.toLowerCase()}/${issue.repository.name.toLowerCase()}#${issue.issue.number}`;
}

function uniqueWinner(
  issues: readonly IssueDetail[],
  field: CompareHighlightField,
  score: (issue: IssueDetail) => number | undefined,
): Partial<Record<CompareHighlightField, string>> {
  const scored = issues
    .map((issue) => {
      const value = score(issue);
      return value === undefined
        ? undefined
        : { issue, key: issueHighlightKey(issue), value };
    })
    .filter(
      (item): item is { issue: IssueDetail; key: string; value: number } =>
        Boolean(item),
    );
  if (scored.length === 0) {
    return {};
  }
  const best = Math.max(...scored.map((item) => item.value));
  const winners = scored.filter((item) => item.value === best);
  if (winners.length !== 1 || scored.every((item) => item.value === best)) {
    return {};
  }
  return { [field]: winners[0]!.key };
}
