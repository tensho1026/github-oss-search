import { profileEndpoints } from "../config/app-config";
import { apiClient, type ApiClient } from "./client";
import type {
  GitHubProfileSnapshotEnvelope,
  GitHubUserEnvelope,
  ProfileAnalysisEnvelope,
} from "./generated";

export function getGitHubUser(
  username: string,
  signal?: AbortSignal,
  client: ApiClient = apiClient,
): Promise<GitHubUserEnvelope> {
  return client.get<GitHubUserEnvelope>(profileEndpoints.user(username), {
    signal,
  });
}

export function getProfileAnalysis(
  username: string,
  signal?: AbortSignal,
  client: ApiClient = apiClient,
): Promise<ProfileAnalysisEnvelope> {
  return client.get<ProfileAnalysisEnvelope>(
    profileEndpoints.analysis(username),
    { signal },
  );
}

export function getProfileSnapshot(
  username: string,
  signal?: AbortSignal,
  client: ApiClient = apiClient,
): Promise<GitHubProfileSnapshotEnvelope> {
  return client.get<GitHubProfileSnapshotEnvelope>(
    profileEndpoints.snapshot(username),
    { signal },
  );
}
