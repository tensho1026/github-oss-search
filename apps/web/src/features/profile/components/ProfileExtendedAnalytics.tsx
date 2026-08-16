import { Badge } from "../../../components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import type {
  EvidenceConfidence,
  EvidenceCount,
  EvidenceStatus,
  ProfileAnalysis,
  ProfileRepositorySample,
  TechnologyEvidence,
} from "../../../shared/api/generated";
import { useI18n, type Locale } from "../../../shared/i18n/i18n-context";
import { formatCompactNumber, formatDate } from "../../../shared/lib/format";
import { ContributionCalendar } from "./ContributionCalendar";

type ProfileExtendedAnalyticsProps = {
  analysis: ProfileAnalysis;
  showPortfolio?: boolean;
};

function enumLabel(value: string, locale: Locale): string {
  const japaneseLabels: Record<string, string> = {
    contributed: "コントリビューション",
    completed: "完了",
    exact: "正確",
    forked: "Fork",
    high: "高",
    in_progress: "進行中",
    locked: "未解放",
    low: "低",
    medium: "中",
    merged_pull_request: "マージ済みPR",
    owned: "所有",
    repository_first: "初回リポジトリ",
    sampled: "サンプル",
    starred: "スター",
    technology_first: "初回技術",
    unavailable: "利用不可",
  };
  if (locale === "ja" && japaneseLabels[value]) {
    return japaneseLabels[value];
  }
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function EvidenceBadge({ status }: { status: EvidenceStatus }) {
  const { locale } = useI18n();
  const tone =
    status === "exact"
      ? "success"
      : status === "sampled"
        ? "warning"
        : "neutral";
  return <Badge variant={tone}>{enumLabel(status, locale)}</Badge>;
}

function ConfidenceBadge({ confidence }: { confidence: EvidenceConfidence }) {
  const { locale } = useI18n();
  const tone =
    confidence === "high"
      ? "success"
      : confidence === "medium"
        ? "warning"
        : "neutral";
  return (
    <Badge variant={tone}>
      {locale === "ja"
        ? `確度${enumLabel(confidence, locale)}`
        : `${enumLabel(confidence, locale)} confidence`}
    </Badge>
  );
}

function EvidenceMetric({
  count,
  label,
}: {
  count: EvidenceCount;
  label: string;
}) {
  const { locale, t } = useI18n();
  return (
    <div className="rounded-xl border border-border bg-muted/30 p-4">
      <div className="flex justify-end">
        <EvidenceBadge status={count.status} />
      </div>
      <p className="mt-4 text-2xl font-semibold tracking-[-0.04em]">
        {count.status === "unavailable"
          ? t("analytics.unavailable")
          : formatCompactNumber(count.value, locale)}
      </p>
      <p className="mt-1 text-xs text-muted-foreground">{label}</p>
    </div>
  );
}

function EvidenceList({ evidence }: { evidence: TechnologyEvidence[] }) {
  const { locale, t } = useI18n();
  if (evidence.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("analytics.noEvidence")}
      </p>
    );
  }
  return (
    <ul className="grid gap-2">
      {evidence.map((item) => (
        <li
          className="flex flex-wrap items-center justify-between gap-2 rounded-lg bg-muted/55 px-3 py-2 text-sm"
          key={`${item.kind}-${item.status}`}
        >
          <span>{enumLabel(item.kind, locale)}</span>
          <span className="flex items-center gap-2">
            <strong>{formatCompactNumber(item.value, locale)}</strong>
            <EvidenceBadge status={item.status} />
          </span>
        </li>
      ))}
    </ul>
  );
}

function RepositorySample({
  label,
  sample,
}: {
  label: string;
  sample: ProfileRepositorySample;
}) {
  const { locale, t } = useI18n();
  return (
    <li className="rounded-xl border border-border bg-muted/30 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="font-semibold">{label}</h3>
        <EvidenceBadge status={sample.status} />
      </div>
      <p className="mt-4 text-2xl font-semibold">
        {sample.status === "unavailable"
          ? t("analytics.unavailable")
          : formatCompactNumber(sample.observed, locale)}
      </p>
      <p className="mt-1 text-xs leading-5 text-muted-foreground">
        {sample.status === "unavailable"
          ? t("analytics.segmentUnavailable")
          : sample.total === null
            ? t("analytics.observedWindow", { limit: sample.limit })
            : t("analytics.totalWindow", {
                limit: sample.limit,
                total: formatCompactNumber(sample.total, locale),
              })}
      </p>
      {sample.status !== "unavailable" ? (
        <p className="mt-3 text-xs text-muted-foreground">
          {t("analytics.activeWindow", {
            count: formatCompactNumber(sample.activeInWindow, locale),
          })}
        </p>
      ) : null}
      {sample.primaryTechnologies.length > 0 ? (
        <ul className="mt-3 flex flex-wrap gap-2">
          {sample.primaryTechnologies.map((technology) => (
            <li key={technology.name}>
              <Badge variant="accent">
                {technology.name} · {technology.percentage}%
              </Badge>
            </li>
          ))}
        </ul>
      ) : null}
    </li>
  );
}

export function ProfileExtendedAnalytics({
  analysis,
  showPortfolio = false,
}: ProfileExtendedAnalyticsProps) {
  const { locale, t } = useI18n();
  const contributions = analysis.contributions;
  const repositorySamples = [
    [enumLabel("owned", locale), analysis.repositoryEvidence.owned],
    [enumLabel("contributed", locale), analysis.repositoryEvidence.contributed],
    [enumLabel("starred", locale), analysis.repositoryEvidence.starred],
    [enumLabel("forked", locale), analysis.repositoryEvidence.forked],
  ] as const;

  return (
    <>
      {showPortfolio ? (
        <ContributionPortfolioPreview analysis={analysis} />
      ) : null}
      <OSSJourneyTimeline analysis={analysis} />
      <ContributionStreakCard analysis={analysis} />
      <OSSQuestCard analysis={analysis} />
      <section aria-labelledby="contribution-activity-heading" className="mt-5">
        <div className="grid gap-5">
          <ContributionCalendar calendar={analysis.contributionCalendar} />
          <Card>
            <CardHeader>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <CardTitle id="contribution-activity-heading">
                    {t("analytics.publicContribution")}
                  </CardTitle>
                  <CardDescription>
                    {t("analytics.window", {
                      from: formatDate(analysis.analysisWindow.from, locale),
                      to: formatDate(analysis.analysisWindow.to, locale),
                    })}
                  </CardDescription>
                </div>
                <Badge variant="neutral">
                  {t("analytics.daysPublic", {
                    days: contributions.windowDays,
                  })}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
              <EvidenceMetric
                count={contributions.commits}
                label={t("analytics.commits")}
              />
              <EvidenceMetric
                count={contributions.pullRequestsOpened}
                label={t("analytics.prs")}
              />
              <EvidenceMetric
                count={contributions.pullRequestReviews}
                label={t("analytics.reviews")}
              />
              <EvidenceMetric
                count={contributions.issuesOpened}
                label={t("analytics.issues")}
              />
              <EvidenceMetric
                count={contributions.repositoriesTouched}
                label={t("analytics.repositories")}
              />
            </CardContent>
          </Card>
        </div>
      </section>

      <section
        aria-label={t("analytics.experienceRecent")}
        className="mt-5 grid gap-5 xl:grid-cols-[0.78fr_1.22fr]"
      >
        <Card>
          <CardHeader>
            <CardTitle className="mt-2">{t("analytics.experience")}</CardTitle>
            <CardDescription>
              {t("analytics.experienceDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap items-center gap-2">
              <Badge
                variant={
                  analysis.ossExperience.level === "unavailable"
                    ? "neutral"
                    : "accent"
                }
              >
                {enumLabel(analysis.ossExperience.level, locale)}
              </Badge>
              <ConfidenceBadge confidence={analysis.ossExperience.confidence} />
            </div>
            <div className="mt-5">
              <EvidenceList evidence={analysis.ossExperience.evidence} />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("analytics.recent")}</CardTitle>
            <CardDescription>
              {t("analytics.recentDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {analysis.recentTechnologies.length > 0 ? (
              <ul className="grid gap-3 sm:grid-cols-2">
                {analysis.recentTechnologies.map((technology) => (
                  <li
                    className="rounded-xl border border-border bg-muted/30 p-4"
                    key={`${technology.kind}-${technology.name}`}
                  >
                    <div className="flex flex-wrap items-start justify-between gap-2">
                      <div>
                        <p className="font-semibold">{technology.name}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {t("analytics.lastUsed", {
                            date: formatDate(technology.lastUsedAt, locale),
                            kind: enumLabel(technology.kind, locale),
                          })}
                        </p>
                      </div>
                      <ConfidenceBadge confidence={technology.confidence} />
                    </div>
                    <p className="mt-3 text-sm text-muted-foreground">
                      {t("analytics.observedRepositories", {
                        count: formatCompactNumber(
                          technology.repositoryCount,
                          locale,
                        ),
                      })}
                    </p>
                    <ul className="mt-3 flex flex-wrap gap-2">
                      {technology.repositorySources.map((source) => (
                        <li key={source}>
                          <Badge variant="neutral">
                            {enumLabel(source, locale)}
                          </Badge>
                        </li>
                      ))}
                    </ul>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-xl bg-muted p-5 text-sm leading-6 text-muted-foreground">
                {t("analytics.noRecent")}
              </p>
            )}
          </CardContent>
        </Card>
      </section>

      <section aria-labelledby="proficiency-heading" className="mt-5">
        <Card>
          <CardHeader>
            <CardTitle id="proficiency-heading">
              {t("analytics.diagnostics")}
            </CardTitle>
            <CardDescription>
              {t("analytics.diagnosticsDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {analysis.proficiency.length > 0 ? (
              <ul className="grid gap-4 lg:grid-cols-2">
                {analysis.proficiency.map((technology) => (
                  <li
                    className="rounded-xl border border-border bg-muted/30 p-4"
                    key={`${technology.kind}-${technology.name}`}
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <p className="font-semibold">{technology.name}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {enumLabel(technology.kind, locale)} ·{" "}
                          {enumLabel(technology.label, locale)}
                        </p>
                      </div>
                      <ConfidenceBadge confidence={technology.confidence} />
                    </div>
                    <div
                      aria-label={`${technology.name} diagnostic level ${technology.level} of 5`}
                      aria-valuemax={5}
                      aria-valuemin={1}
                      aria-valuenow={technology.level}
                      className="mt-4 h-2.5 overflow-hidden rounded-full bg-muted-strong"
                      role="progressbar"
                    >
                      <div
                        className="h-full rounded-full bg-accent"
                        style={{ width: `${technology.level * 20}%` }}
                      />
                    </div>
                    <p className="mt-2 text-xs text-muted-foreground">
                      {t("analytics.levelScore", {
                        level: technology.level,
                        score: technology.score,
                      })}
                    </p>
                    <div className="mt-4">
                      <EvidenceList evidence={technology.evidence} />
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-xl bg-muted p-5 text-sm leading-6 text-muted-foreground">
                {t("analytics.noDiagnostics")}
              </p>
            )}
          </CardContent>
        </Card>
      </section>

      <section aria-labelledby="repository-sources-heading" className="mt-5">
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <CardTitle id="repository-sources-heading">
                  {t("analytics.sourceEvidence")}
                </CardTitle>
                <CardDescription>
                  {t("analytics.sourceDescription")}
                </CardDescription>
              </div>
              <Badge
                variant={
                  analysis.languageStatus === "exact" ? "success" : "warning"
                }
              >
                {t("analytics.languageEvidence", {
                  status: enumLabel(analysis.languageStatus, locale),
                })}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <ul className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              {repositorySamples.map(([label, sample]) => (
                <RepositorySample key={label} label={label} sample={sample} />
              ))}
            </ul>
            <p className="mt-4 text-xs text-muted-foreground">
              {t("analytics.starredNote")}
            </p>
          </CardContent>
        </Card>
      </section>
    </>
  );
}

function OSSQuestCard({ analysis }: { analysis: ProfileAnalysis }) {
  const { locale, t } = useI18n();
  const quest = analysis.ossQuest;
  const next = quest.items.find((item) => item.id === quest.nextQuestId);
  const title = (id: string, fallback: string) => {
    switch (id) {
      case "first_issue_comment":
        return t("analytics.quest.first_issue_comment.title");
      case "first_pr":
        return t("analytics.quest.first_pr.title");
      case "first_review":
        return t("analytics.quest.first_review.title");
      case "first_merge":
        return t("analytics.quest.first_merge.title");
      case "three_repositories":
        return t("analytics.quest.three_repositories.title");
      default:
        return fallback;
    }
  };
  const nextAction = (id: string, fallback: string) => {
    switch (id) {
      case "first_issue_comment":
        return t("analytics.quest.first_issue_comment.action");
      case "first_pr":
        return t("analytics.quest.first_pr.action");
      case "first_review":
        return t("analytics.quest.first_review.action");
      case "first_merge":
        return t("analytics.quest.first_merge.action");
      case "three_repositories":
        return t("analytics.quest.three_repositories.action");
      default:
        return fallback;
    }
  };
  return (
    <section aria-labelledby="oss-quest-heading" className="mt-5">
      <Card>
        <CardHeader>
          <CardTitle id="oss-quest-heading">
            {t("analytics.questTitle")}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ol className="grid gap-2">
            {quest.items.map((item) => (
              <li key={item.id}>
                <div className="flex items-center justify-between gap-3">
                  <span className="font-semibold">
                    {title(item.id, item.title)}
                  </span>
                  <Badge
                    variant={
                      item.status === "completed" ? "success" : "neutral"
                    }
                  >
                    {enumLabel(item.status, locale)} {item.current}/
                    {item.target}
                  </Badge>
                </div>
              </li>
            ))}
          </ol>
          {next ? (
            <p className="mt-3">
              {t("analytics.questNext", {
                action: nextAction(next.id, next.nextAction),
              })}
            </p>
          ) : null}
        </CardContent>
      </Card>
    </section>
  );
}

function ContributionStreakCard({ analysis }: { analysis: ProfileAnalysis }) {
  const { locale, t } = useI18n();
  const streak = analysis.contributionStreak;
  return (
    <section aria-labelledby="contribution-streak-heading" className="mt-5">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle id="contribution-streak-heading">
                {t("analytics.streakTitle")}
              </CardTitle>
            </div>
            <EvidenceBadge status={streak.status} />
          </div>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-3 gap-3">
            {[
              [t("analytics.current"), streak.currentWeeks],
              [t("analytics.longest"), streak.longestWeeks],
              [t("analytics.activeWeeks"), streak.qualifyingWeeks],
            ].map(([label, value]) => (
              <div className="rounded-xl bg-muted p-3" key={label}>
                <dt className="text-xs text-muted-foreground">{label}</dt>
                <dd className="mt-1 text-xl font-semibold">{value}</dd>
              </div>
            ))}
          </dl>
          {streak.status === "unavailable" ? (
            <p className="mt-4 text-sm text-muted-foreground">
              {t("analytics.streakUnavailable")}
            </p>
          ) : (
            <ul
              className="mt-4 grid gap-2"
              aria-label={t("analytics.qualifyingWeeks")}
            >
              {streak.weeks.map((week) => (
                <li
                  className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border px-3 py-2"
                  key={week.startedAt}
                >
                  <span className="text-sm">
                    {t("analytics.verifiedEvents", {
                      count: formatCompactNumber(week.eventCount, locale),
                      date: formatDate(week.startedAt, locale),
                    })}
                  </span>
                  <a
                    className="text-accent hover:underline"
                    href={week.evidenceUrls[0]}
                    rel="noreferrer"
                    target="_blank"
                  >
                    {t("analytics.evidence")}
                  </a>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </section>
  );
}

function OSSJourneyTimeline({ analysis }: { analysis: ProfileAnalysis }) {
  const { locale, t } = useI18n();
  const journey = analysis.ossJourney;
  const milestoneTitle = (milestone: (typeof journey.milestones)[number]) => {
    if (milestone.kind === "merged_pull_request") {
      return t("analytics.journeyMergedTitle", {
        pullRequest: milestone.id.match(/#\d+$/)?.[0] ?? "",
        repository: milestone.repositoryName,
      });
    }
    if (milestone.kind === "repository_first") {
      return t("analytics.journeyRepositoryTitle", {
        repository: milestone.repositoryName,
      });
    }
    if (milestone.kind === "technology_first") {
      return t("analytics.journeyTechnologyTitle", {
        technology: milestone.technology ?? "",
      });
    }
    return milestone.title;
  };
  const milestoneDescription = (
    milestone: (typeof journey.milestones)[number],
  ) => {
    if (milestone.kind === "merged_pull_request") {
      return t("analytics.journeyMergedDescription", {
        title: milestone.description.replace(/^Observed public merge: /, ""),
      });
    }
    if (milestone.kind === "repository_first") {
      return t("analytics.journeyRepositoryDescription");
    }
    if (milestone.kind === "technology_first") {
      return t("analytics.journeyTechnologyDescription");
    }
    return milestone.description;
  };
  return (
    <section aria-labelledby="oss-journey-heading" className="mt-5">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle id="oss-journey-heading">
                {t("analytics.journeyTitle")}
              </CardTitle>
              <CardDescription>
                {t("analytics.journeyDescription")}
              </CardDescription>
            </div>
            <EvidenceBadge status={journey.status} />
          </div>
        </CardHeader>
        <CardContent>
          {journey.status === "unavailable" ? (
            <p className="text-sm text-muted-foreground">
              {t("analytics.journeyUnavailable")}
            </p>
          ) : journey.milestones.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {t("analytics.journeyEmpty")}
            </p>
          ) : (
            <ol className="relative grid gap-4 border-l border-border pl-5">
              {journey.milestones.map((milestone) => (
                <li className="relative" key={milestone.id}>
                  <span
                    aria-hidden="true"
                    className="absolute -left-[1.49rem] top-1.5 size-2 rounded-full bg-accent ring-4 ring-surface"
                  />
                  <div className="flex flex-wrap items-center gap-2">
                    <time
                      className="font-mono text-xs text-muted-foreground"
                      dateTime={milestone.occurredAt}
                    >
                      {formatDate(milestone.occurredAt, locale)}
                    </time>
                    <Badge variant="neutral">
                      {enumLabel(milestone.kind, locale)}
                    </Badge>
                  </div>
                  <p className="mt-1 font-semibold">
                    {milestoneTitle(milestone)}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {milestoneDescription(milestone)}
                  </p>
                  <a
                    className="mt-2 inline-flex rounded-md text-sm font-semibold text-accent outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
                    href={milestone.evidenceUrl}
                    rel="noreferrer"
                    target="_blank"
                  >
                    {t("analytics.viewEvidence")}
                  </a>
                </li>
              ))}
            </ol>
          )}
          <p className="mt-4 text-xs text-muted-foreground">
            {t("analytics.journeyAnalyzed", {
              date: formatDate(journey.analyzedAt, locale),
            })}
          </p>
        </CardContent>
      </Card>
    </section>
  );
}

function ContributionPortfolioPreview({
  analysis,
}: {
  analysis: ProfileAnalysis;
}) {
  const { locale, t } = useI18n();
  const portfolio = analysis.contributionPortfolio;
  return (
    <section aria-labelledby="portfolio-heading" className="mt-5">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle id="portfolio-heading">
                {t("analytics.portfolioTitle")}
              </CardTitle>
              <CardDescription>
                {t("analytics.portfolioDescription")}
              </CardDescription>
            </div>
            <EvidenceBadge status={portfolio.status} />
          </div>
        </CardHeader>
        <CardContent>
          {portfolio.status === "unavailable" ? (
            <p className="text-sm text-muted-foreground">
              {t("analytics.portfolioUnavailable")}
            </p>
          ) : (
            <>
              <dl className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {[
                  [t("analytics.mergedPrs"), portfolio.totalMerged],
                  [t("analytics.displayed"), portfolio.displayedMerged],
                  [t("analytics.repositories"), portfolio.repositoryCount],
                  [t("analytics.languages"), portfolio.languages.length],
                ].map(([label, value]) => (
                  <div className="rounded-xl bg-muted p-3" key={label}>
                    <dt className="text-xs text-muted-foreground">{label}</dt>
                    <dd className="mt-1 text-xl font-semibold">{value}</dd>
                  </div>
                ))}
              </dl>
              <ul
                className="mt-4 flex flex-wrap gap-2"
                aria-label={t("analytics.portfolioLanguages")}
              >
                {portfolio.languages.map((language) => (
                  <li key={language.name}>
                    <Badge variant="accent">
                      {language.name} {language.count}
                    </Badge>
                  </li>
                ))}
              </ul>
              <ul className="mt-4 grid gap-3">
                {portfolio.contributions.map((item) => (
                  <li
                    className="rounded-xl border border-border bg-muted/30 p-4"
                    key={`${item.repositoryOwner}/${item.repositoryName}#${item.number}`}
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <p className="font-mono text-xs text-muted-foreground">
                          {item.repositoryOwner}/{item.repositoryName}#
                          {item.number}
                        </p>
                        <p className="mt-1 font-semibold">{item.title}</p>
                      </div>
                      <Badge variant="success">
                        {t("analytics.observedMerged")}
                      </Badge>
                    </div>
                    <p className="mt-2 text-sm text-muted-foreground">
                      {t("analytics.portfolioContributionSummary", {
                        date: formatDate(item.mergedAt, locale),
                        repository: `${item.repositoryOwner}/${item.repositoryName}`,
                      })}
                    </p>
                    <a
                      className="mt-3 inline-flex rounded-md text-sm font-semibold text-accent outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
                      href={item.url}
                      rel="noreferrer"
                      target="_blank"
                    >
                      {t("analytics.viewCanonicalPr")}
                    </a>
                  </li>
                ))}
              </ul>
              <p className="mt-4 text-xs text-muted-foreground">
                {t("analytics.portfolioAnalyzed", {
                  date: formatDate(portfolio.analyzedAt, locale),
                })}
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
