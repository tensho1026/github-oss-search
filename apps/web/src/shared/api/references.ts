import { apiClient } from "./client";
import type {
  GitHubReferenceObservationEnvelope,
  GitHubReferenceObservationRequest,
} from "./generated";

export function observeGitHubReference(
  request: GitHubReferenceObservationRequest,
): Promise<GitHubReferenceObservationEnvelope> {
  return apiClient.post<
    GitHubReferenceObservationEnvelope,
    GitHubReferenceObservationRequest
  >("/api/github/references/observe", request);
}
