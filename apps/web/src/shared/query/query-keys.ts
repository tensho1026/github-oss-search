export const queryKeys = Object.freeze({
  account: Object.freeze({
    bookmarks: ["account", "bookmarks"] as const,
    bookmarkPage(page: number, perPage: number) {
      return ["account", "bookmarks", page, perPage] as const;
    },
    issueClaims: ["account", "issue-claims"] as const,
    issueClaimPage(filter: string, page: number, perPage: number) {
      return ["account", "issue-claims", filter, page, perPage] as const;
    },
    preferences: ["account", "preferences"] as const,
    profileSnapshots: ["account", "profile-snapshots"] as const,
    root: ["account"] as const,
    savedSearches: ["account", "saved-searches"] as const,
    savedSearchPage(page: number, perPage: number) {
      return ["account", "saved-searches", page, perPage] as const;
    },
  }),
  auth: Object.freeze({
    session: ["auth", "session"] as const,
  }),
  issues: Object.freeze({
    detail(
      owner: string,
      repository: string,
      issueNumber: number,
      skills: readonly string[],
    ) {
      return [
        "issues",
        "detail",
        owner.toLowerCase(),
        repository.toLowerCase(),
        issueNumber,
        [...skills],
      ] as const;
    },
    root: ["issues"] as const,
    search(canonicalSearch: string) {
      return ["issues", "search", canonicalSearch] as const;
    },
  }),
  profile: Object.freeze({
    analysis(username: string) {
      return ["profile", username.toLowerCase(), "analysis"] as const;
    },
    root: ["profile"] as const,
    snapshot(username: string) {
      return ["profile", username.toLowerCase(), "snapshot"] as const;
    },
    user(username: string) {
      return ["profile", username.toLowerCase(), "user"] as const;
    },
  }),
  repositories: Object.freeze({
    root: ["repositories"] as const,
    search(canonicalSearch: string) {
      return ["repositories", "search", canonicalSearch] as const;
    },
  }),
});
