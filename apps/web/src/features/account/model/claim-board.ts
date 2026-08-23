import type {
  IssueClaim,
  IssueClaimStatus,
  IssueClaimUpdateRequest,
} from "../../../shared/api/generated";

export type ClaimMove =
  | { ok: true; request: IssueClaimUpdateRequest }
  | { ok: false; reason: "unchanged" | "pull_request_required" };

export function claimMoveRequest(
  claim: IssueClaim,
  status: IssueClaimStatus,
): ClaimMove {
  if (claim.status === status) return { ok: false, reason: "unchanged" };
  if (
    (status === "pr_submitted" || status === "merged") &&
    !claim.pullRequest
  ) {
    return { ok: false, reason: "pull_request_required" };
  }
  return {
    ok: true,
    request: {
      archived: claim.archived,
      pullRequest: claim.pullRequest,
      status,
      version: claim.version,
    },
  };
}
