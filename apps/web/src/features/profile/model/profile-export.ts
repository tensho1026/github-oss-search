import type {
  GitHubUser,
  ProfileAnalysis,
} from "../../../shared/api/generated";

export function profileMarkdown(
  user: GitHubUser,
  analysis: ProfileAnalysis,
): string {
  const technologies = [
    ...analysis.languages.map(({ name }) => name),
    ...analysis.frameworks,
  ].slice(0, 10);
  const journey = analysis.ossJourney.milestones.slice(-5);
  const contributions = analysis.contributionPortfolio.contributions.slice(
    0,
    5,
  );
  return [
    `# ${user.name || user.login} — OSS profile`,
    "",
    `GitHub: [@${user.login}](https://github.com/${encodeURIComponent(user.login)})`,
    "",
    "## Technologies",
    "",
    technologies.length > 0
      ? technologies.map((value) => `- ${value}`).join("\n")
      : "- No public evidence available",
    "",
    `## OSS Journey (${analysis.ossJourney.status})`,
    "",
    journey.length > 0
      ? journey
          .map(
            (item) =>
              `- ${item.occurredAt.slice(0, 10)} — [${item.title}](${item.evidenceUrl})`,
          )
          .join("\n")
      : "- No bounded milestone available",
    "",
    `## Merged pull requests (${analysis.contributionPortfolio.totalMerged} observed)`,
    "",
    contributions.length > 0
      ? contributions
          .map(
            (item) =>
              `- [${item.repositoryOwner}/${item.repositoryName}#${item.number}: ${item.title}](${item.url})`,
          )
          .join("\n")
      : "- No bounded merged pull request available",
    "",
    `Current verified weekly streak: ${analysis.contributionStreak.currentWeeks} week(s)`,
    "",
    "_Generated from bounded public GitHub evidence by IssueScout. Sampled evidence is not a complete employment or contribution history._",
    "",
  ].join("\n");
}

export function profileCardLines(user: GitHubUser, analysis: ProfileAnalysis) {
  return {
    heading: user.name || user.login,
    login: `@${user.login}`,
    merged: `${analysis.contributionPortfolio.totalMerged} merged PRs observed`,
    streak: `${analysis.contributionStreak.currentWeeks} week current OSS streak`,
    technologies: [
      ...analysis.languages.map(({ name }) => name),
      ...analysis.frameworks,
    ].slice(0, 8),
    journey: analysis.ossJourney.milestones.slice(-3).map(({ title }) => title),
  };
}
