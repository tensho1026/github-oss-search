import type {
  GitHubUserEnvelope,
  ProfileAnalysisEnvelope,
} from "../shared/api/generated";

const meta = {
  rateLimitRemaining: 4_992,
  requestId: "req_profile_test",
  timestamp: "2026-07-30T00:00:00Z",
};

export const gitHubUserFixture: GitHubUserEnvelope = {
  data: {
    avatarUrl: "https://avatars.githubusercontent.com/u/1?v=4",
    bio: "Builds useful developer tools.",
    followers: 1_250,
    following: 42,
    login: "octocat",
    name: "The Octocat",
    publicRepos: 8,
    repositories: [
      {
        defaultBranch: "main",
        description: "A typed service",
        forks: 3,
        fullName: "octocat/typed-service",
        isArchived: false,
        isFork: false,
        mainLanguage: "TypeScript",
        name: "typed-service",
        openIssues: 4,
        owner: "octocat",
        pushedAt: "2026-07-29T00:00:00Z",
        stars: 120,
        updatedAt: "2026-07-29T00:00:00Z",
        url: "https://github.com/octocat/typed-service",
      },
    ],
  },
  meta,
};

export const profileAnalysisFixture: ProfileAnalysisEnvelope = {
  data: {
    analysisWindow: {
      days: 365,
      from: "2025-07-30T00:00:00Z",
      publicOnly: true,
      to: "2026-07-30T00:00:00Z",
    },
    contributions: {
      commits: { status: "sampled", value: 18 },
      issuesOpened: { status: "exact", value: 3 },
      pullRequestReviews: { status: "sampled", value: 4 },
      pullRequestsOpened: { status: "exact", value: 7 },
      repositoriesTouched: { status: "sampled", value: 1 },
      windowDays: 365,
    },
    contributionCalendar: {
      from: "2026-07-19",
      status: "exact",
      to: "2026-07-25",
      total: 3,
      weeks: [
        {
          days: [
            { count: 0, date: "2026-07-19", level: "none", weekday: 0 },
            {
              count: 1,
              date: "2026-07-20",
              level: "first_quartile",
              weekday: 1,
            },
            { count: 0, date: "2026-07-21", level: "none", weekday: 2 },
            {
              count: 2,
              date: "2026-07-22",
              level: "second_quartile",
              weekday: 3,
            },
            { count: 0, date: "2026-07-23", level: "none", weekday: 4 },
            { count: 0, date: "2026-07-24", level: "none", weekday: 5 },
            { count: 0, date: "2026-07-25", level: "none", weekday: 6 },
          ],
          firstDay: "2026-07-19",
          index: 0,
        },
      ],
    },
    contributionPortfolio: {
      analyzedAt: "2026-07-30T00:00:00Z",
      contributions: [
        {
          language: "Go",
          mergedAt: "2026-07-20T12:00:00Z",
          number: 42,
          repositoryName: "project",
          repositoryOwner: "community",
          summary:
            "Merged public pull request in community/project: Add bounded retries",
          title: "Add bounded retries",
          url: "https://github.com/community/project/pull/42",
        },
      ],
      displayedMerged: 1,
      hasMore: true,
      languages: [{ count: 1, name: "Go" }],
      repositoryCount: 1,
      status: "sampled",
      totalMerged: 2,
    },
    contributionStreak: {
      analyzedAt: "2026-07-30T00:00:00Z",
      currentWeeks: 0,
      longestWeeks: 1,
      qualifyingWeeks: 1,
      status: "sampled",
      timezone: "UTC",
      weeks: [
        {
          endedAt: "2026-07-26T23:59:59.999999999Z",
          eventCount: 1,
          evidenceUrls: ["https://github.com/community/project/pull/42"],
          startedAt: "2026-07-20T00:00:00Z",
        },
      ],
      weekStartsOn: "monday",
    },
    ossJourney: {
      analyzedAt: "2026-07-30T00:00:00Z",
      milestones: [
        {
          description: "Observed public merge: Add bounded retries",
          evidenceUrl: "https://github.com/community/project/pull/42",
          id: "merged:community/project#42",
          kind: "merged_pull_request",
          occurredAt: "2026-07-20T12:00:00Z",
          repositoryName: "community/project",
          title: "Merged PR #42 in community/project",
        },
        {
          description:
            "Earliest merged PR for this repository in the bounded sample.",
          evidenceUrl: "https://github.com/community/project/pull/42",
          id: "repository:community/project",
          kind: "repository_first",
          occurredAt: "2026-07-20T12:00:00Z",
          repositoryName: "community/project",
          title: "First observed contribution to community/project",
        },
        {
          description:
            "Earliest merged PR using this repository's primary language in the bounded sample.",
          evidenceUrl: "https://github.com/community/project/pull/42",
          id: "technology:go",
          kind: "technology_first",
          occurredAt: "2026-07-20T12:00:00Z",
          repositoryName: "community/project",
          technology: "Go",
          title: "First observed Go contribution",
        },
      ],
      status: "sampled",
    },
    ossQuest: {
      catalogVersion: "2026-08-01",
      completed: 2,
      evaluatedAt: "2026-07-30T00:00:00Z",
      items: [
        {
          completedAt: null,
          current: 0,
          description: "Join a public issue discussion.",
          id: "first_issue_comment",
          nextAction: "Comment on an issue after confirming you can help.",
          status: "unavailable",
          target: 1,
          title: "Comment on your first issue",
        },
        {
          completedAt: null,
          current: 1,
          description: "Create a public OSS pull request.",
          id: "first_pr",
          nextAction: "Find a matching issue and open a focused PR.",
          status: "completed",
          target: 1,
          title: "Open your first pull request",
        },
        {
          completedAt: null,
          current: 1,
          description: "Review a public OSS pull request.",
          id: "first_review",
          nextAction: "Review an open PR and leave actionable feedback.",
          status: "completed",
          target: 1,
          title: "Complete your first review",
        },
        {
          completedAt: null,
          current: 0,
          description: "Have a public OSS pull request merged.",
          id: "first_merge",
          nextAction: "Respond to review feedback on an open pull request.",
          status: "in_progress",
          target: 1,
          title: "Get your first PR merged",
        },
        {
          completedAt: null,
          current: 1,
          description:
            "Build verified merged contributions across three public projects.",
          id: "three_repositories",
          nextAction: "Find a suitable issue in a new repository.",
          status: "locked",
          target: 3,
          title: "Contribute to 3 repositories",
        },
      ],
      nextQuestId: "first_merge",
      total: 5,
    },
    frameworks: ["React", "Gin"],
    languages: [
      { name: "TypeScript", percentage: 65 },
      { name: "Go", percentage: 35 },
    ],
    languageStatus: "exact",
    ossExperience: {
      confidence: "high",
      evidence: [
        {
          kind: "authored_pull_requests",
          status: "exact",
          value: 7,
        },
      ],
      level: "active",
      publicOnly: true,
    },
    proficiency: [
      {
        confidence: "medium",
        evidence: [
          {
            kind: "language_share_percentage",
            status: "exact",
            value: 65,
          },
        ],
        kind: "language",
        label: "developing",
        level: 2,
        name: "TypeScript",
        score: 39,
      },
    ],
    recentTechnologies: [
      {
        confidence: "medium",
        kind: "language",
        lastUsedAt: "2026-07-29T00:00:00Z",
        name: "TypeScript",
        repositoryCount: 2,
        repositorySources: ["contributed", "owned"],
      },
    ],
    repositoryEvidence: {
      contributed: {
        activeInWindow: 1,
        limit: 20,
        observed: 1,
        primaryTechnologies: [{ name: "TypeScript", percentage: 100 }],
        status: "exact",
        total: 1,
      },
      forked: {
        activeInWindow: 1,
        limit: 20,
        observed: 1,
        primaryTechnologies: [{ name: "Go", percentage: 100 }],
        status: "exact",
        total: 1,
      },
      owned: {
        activeInWindow: 8,
        limit: 20,
        observed: 8,
        primaryTechnologies: [
          { name: "TypeScript", percentage: 65 },
          { name: "Go", percentage: 35 },
        ],
        status: "exact",
        total: 8,
      },
      starred: {
        activeInWindow: 1,
        limit: 20,
        observed: 1,
        primaryTechnologies: [{ name: "Rust", percentage: 100 }],
        status: "sampled",
        total: null,
      },
    },
    repositoriesAnalyzed: 8,
    username: "octocat",
    warnings: [],
  },
  meta,
};

export function errorEnvelope(status: 404 | 429) {
  return {
    error: {
      code:
        status === 404 ? "GITHUB_USER_NOT_FOUND" : "GITHUB_RATE_LIMIT_EXCEEDED",
      message: status === 404 ? "GitHub user was not found" : "Rate limited",
    },
    meta,
  };
}
