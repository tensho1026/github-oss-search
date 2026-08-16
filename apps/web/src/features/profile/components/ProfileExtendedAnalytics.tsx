import { Activity, GitPullRequest, Radar, Star } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { Icon } from "../../../components/ui/icon";
import type {
  EvidenceConfidence,
  EvidenceCount,
  EvidenceStatus,
  ProfileAnalysis,
  ProfileRepositorySample,
  TechnologyEvidence,
} from "../../../shared/api/generated";
import { formatCompactNumber, formatDate } from "../../../shared/lib/format";
import { ContributionCalendar } from "./ContributionCalendar";

type ProfileExtendedAnalyticsProps = {
  analysis: ProfileAnalysis;
  showPortfolio?: boolean;
};

function enumLabel(value: string): string {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function EvidenceBadge({ status }: { status: EvidenceStatus }) {
  const tone =
    status === "exact"
      ? "success"
      : status === "sampled"
        ? "warning"
        : "neutral";
  return <Badge variant={tone}>{enumLabel(status)}</Badge>;
}

function ConfidenceBadge({ confidence }: { confidence: EvidenceConfidence }) {
  const tone =
    confidence === "high"
      ? "success"
      : confidence === "medium"
        ? "warning"
        : "neutral";
  return <Badge variant={tone}>{enumLabel(confidence)} confidence</Badge>;
}

function EvidenceMetric({
  count,
  icon,
  label,
}: {
  count: EvidenceCount;
  icon: typeof Activity;
  label: string;
}) {
  return (
    <div className="rounded-xl border border-border bg-muted/30 p-4">
      <div className="flex items-start justify-between gap-2">
        <span className="grid size-9 place-items-center rounded-lg bg-surface text-accent">
          <Icon icon={icon} />
        </span>
        <EvidenceBadge status={count.status} />
      </div>
      <p className="mt-4 text-2xl font-semibold tracking-[-0.04em]">
        {count.status === "unavailable"
          ? "Unavailable"
          : formatCompactNumber(count.value)}
      </p>
      <p className="mt-1 text-xs text-muted-foreground">{label}</p>
    </div>
  );
}

function EvidenceList({ evidence }: { evidence: TechnologyEvidence[] }) {
  if (evidence.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No bounded public evidence was returned.
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
          <span>{enumLabel(item.kind)}</span>
          <span className="flex items-center gap-2">
            <strong>{formatCompactNumber(item.value)}</strong>
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
  return (
    <li className="rounded-xl border border-border bg-muted/30 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="font-semibold">{label}</h3>
        <EvidenceBadge status={sample.status} />
      </div>
      <p className="mt-4 text-2xl font-semibold">
        {sample.status === "unavailable"
          ? "Unavailable"
          : formatCompactNumber(sample.observed)}
      </p>
      <p className="mt-1 text-xs leading-5 text-muted-foreground">
        {sample.status === "unavailable"
          ? "GitHub did not provide this public segment."
          : sample.total === null
            ? `Observed in a bounded window of up to ${sample.limit}.`
            : `${formatCompactNumber(sample.total)} public total · bounded limit ${sample.limit}.`}
      </p>
      {sample.status !== "unavailable" ? (
        <p className="mt-3 text-xs text-muted-foreground">
          {formatCompactNumber(sample.activeInWindow)} active during the
          analysis window
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
  const contributions = analysis.contributions;
  const repositorySamples = [
    ["Owned", analysis.repositoryEvidence.owned],
    ["Contributed", analysis.repositoryEvidence.contributed],
    ["Starred", analysis.repositoryEvidence.starred],
    ["Forked", analysis.repositoryEvidence.forked],
  ] as const;

  return (
    <>
      {showPortfolio ? (
        <ContributionPortfolioPreview analysis={analysis} />
      ) : null}
      <OSSJourneyTimeline analysis={analysis} />
      <ContributionStreakCard analysis={analysis} />
      <section aria-labelledby="contribution-activity-heading" className="mt-5">
        <div className="grid gap-5">
          <ContributionCalendar calendar={analysis.contributionCalendar} />
          <Card>
            <CardHeader>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <CardTitle id="contribution-activity-heading">
                    Public contribution activity
                  </CardTitle>
                  <CardDescription>
                    Evidence from {formatDate(analysis.analysisWindow.from)} to{" "}
                    {formatDate(analysis.analysisWindow.to)}. Sampled counts are
                    observations, never presented as lifetime totals.
                  </CardDescription>
                </div>
                <Badge variant="neutral">
                  {contributions.windowDays}-day window · public only
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
              <EvidenceMetric
                count={contributions.commits}
                icon={Activity}
                label="Observed commits"
              />
              <EvidenceMetric
                count={contributions.pullRequestsOpened}
                icon={GitPullRequest}
                label="Pull requests opened"
              />
              <EvidenceMetric
                count={contributions.pullRequestReviews}
                icon={Activity}
                label="Observed PR reviews"
              />
              <EvidenceMetric
                count={contributions.issuesOpened}
                icon={Activity}
                label="Issues opened"
              />
              <EvidenceMetric
                count={contributions.repositoriesTouched}
                icon={Activity}
                label="Repositories touched"
              />
            </CardContent>
          </Card>
        </div>
      </section>

      <section
        aria-label="OSS experience and recent technologies"
        className="mt-5 grid gap-5 xl:grid-cols-[0.78fr_1.22fr]"
      >
        <Card>
          <CardHeader>
            <span className="grid size-11 place-items-center rounded-xl bg-accent-soft text-accent-soft-foreground">
              <Icon icon={Radar} />
            </span>
            <CardTitle className="mt-2">OSS experience signal</CardTitle>
            <CardDescription>
              Rule-based server summary of public contribution evidence, not a
              certification.
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
                {enumLabel(analysis.ossExperience.level)}
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
            <CardTitle>Recently observed technologies</CardTitle>
            <CardDescription>
              Last use, repository sources, count, and confidence from bounded
              public repository evidence.
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
                          {enumLabel(technology.kind)} · last used{" "}
                          {formatDate(technology.lastUsedAt)}
                        </p>
                      </div>
                      <ConfidenceBadge confidence={technology.confidence} />
                    </div>
                    <p className="mt-3 text-sm text-muted-foreground">
                      {formatCompactNumber(technology.repositoryCount)} observed
                      repositories
                    </p>
                    <ul className="mt-3 flex flex-wrap gap-2">
                      {technology.repositorySources.map((source) => (
                        <li key={source}>
                          <Badge variant="neutral">{enumLabel(source)}</Badge>
                        </li>
                      ))}
                    </ul>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-xl bg-muted p-5 text-sm leading-6 text-muted-foreground">
                No recently observed technology evidence was available in this
                public window.
              </p>
            )}
          </CardContent>
        </Card>
      </section>

      <section aria-labelledby="proficiency-heading" className="mt-5">
        <Card>
          <CardHeader>
            <CardTitle id="proficiency-heading">
              Five-level technology diagnostics
            </CardTitle>
            <CardDescription>
              Server-assigned levels render their exact inputs and confidence.
              They do not claim comprehensive skill measurement.
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
                          {enumLabel(technology.kind)} ·{" "}
                          {enumLabel(technology.label)}
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
                      Level {technology.level}/5 · evidence score{" "}
                      {technology.score}/100
                    </p>
                    <div className="mt-4">
                      <EvidenceList evidence={technology.evidence} />
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="rounded-xl bg-muted p-5 text-sm leading-6 text-muted-foreground">
                No five-level diagnostic could be supported by the available
                public evidence.
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
                  Repository source evidence
                </CardTitle>
                <CardDescription>
                  Owned, contributed, starred, and forked public observations
                  remain separate so unavailable data never looks like zero.
                </CardDescription>
              </div>
              <Badge
                variant={
                  analysis.languageStatus === "exact" ? "success" : "warning"
                }
              >
                Language evidence · {enumLabel(analysis.languageStatus)}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <ul className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              {repositorySamples.map(([label, sample]) => (
                <RepositorySample key={label} label={label} sample={sample} />
              ))}
            </ul>
            <p className="mt-4 flex items-center gap-2 text-xs text-muted-foreground">
              <Icon icon={Star} />
              Starred totals can be privacy-ambiguous for the API viewer and may
              intentionally omit a total.
            </p>
          </CardContent>
        </Card>
      </section>
    </>
  );
}

function ContributionStreakCard({ analysis }: { analysis: ProfileAnalysis }) {
  const streak = analysis.contributionStreak;
  return (
    <section aria-labelledby="contribution-streak-heading" className="mt-5">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle id="contribution-streak-heading">
                Contribution Streak
              </CardTitle>
            </div>
            <EvidenceBadge status={streak.status} />
          </div>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-3 gap-3">
            {[
              ["Current", streak.currentWeeks],
              ["Longest", streak.longestWeeks],
              ["Active weeks", streak.qualifyingWeeks],
            ].map(([label, value]) => (
              <div className="rounded-xl bg-muted p-3" key={label}>
                <dt className="text-xs text-muted-foreground">{label}</dt>
                <dd className="mt-1 text-xl font-semibold">{value}</dd>
              </div>
            ))}
          </dl>
          {streak.status === "unavailable" ? (
            <p className="mt-4 text-sm text-muted-foreground">
              Weekly contribution evidence is unavailable.
            </p>
          ) : (
            <ul className="mt-4 grid gap-2" aria-label="Qualifying weeks">
              {streak.weeks.map((week) => (
                <li
                  className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border px-3 py-2"
                  key={week.startedAt}
                >
                  <span className="text-sm">
                    {formatDate(week.startedAt)} · {week.eventCount} verified
                    event{week.eventCount === 1 ? "" : "s"}
                  </span>
                  <a
                    className="text-accent hover:underline"
                    href={week.evidenceUrls[0]}
                    rel="noreferrer"
                    target="_blank"
                  >
                    Evidence
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
  const journey = analysis.ossJourney;
  return (
    <section aria-labelledby="oss-journey-heading" className="mt-5">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle id="oss-journey-heading">OSS Journey</CardTitle>
              <CardDescription>
                A chronological timeline backed by canonical public GitHub
                evidence. First means earliest in the observed sample.
              </CardDescription>
            </div>
            <EvidenceBadge status={journey.status} />
          </div>
        </CardHeader>
        <CardContent>
          {journey.status === "unavailable" ? (
            <p className="text-sm text-muted-foreground">
              No verified public journey evidence is available.
            </p>
          ) : journey.milestones.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No milestone was observed in this bounded sample.
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
                      {formatDate(milestone.occurredAt)}
                    </time>
                    <Badge variant="neutral">{enumLabel(milestone.kind)}</Badge>
                  </div>
                  <p className="mt-1 font-semibold">{milestone.title}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {milestone.description}
                  </p>
                  <a
                    className="mt-2 inline-flex rounded-md text-sm font-semibold text-accent outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
                    href={milestone.evidenceUrl}
                    rel="noreferrer"
                    target="_blank"
                  >
                    View evidence
                  </a>
                </li>
              ))}
            </ol>
          )}
          <p className="mt-4 text-xs text-muted-foreground">
            Analyzed {formatDate(journey.analyzedAt)}. Ordering uses normalized
            UTC timestamps and stable milestone IDs.
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
  const portfolio = analysis.contributionPortfolio;
  return (
    <section aria-labelledby="portfolio-heading" className="mt-5">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <CardTitle id="portfolio-heading">
                Private contribution portfolio preview
              </CardTitle>
              <CardDescription>
                Bounded public merged-PR facts for your account. Nothing here is
                published automatically.
              </CardDescription>
            </div>
            <EvidenceBadge status={portfolio.status} />
          </div>
        </CardHeader>
        <CardContent>
          {portfolio.status === "unavailable" ? (
            <p className="text-sm text-muted-foreground">
              Public merged pull-request evidence is unavailable.
            </p>
          ) : (
            <>
              <dl className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {[
                  ["Merged PRs", portfolio.totalMerged],
                  ["Displayed", portfolio.displayedMerged],
                  ["Repositories", portfolio.repositoryCount],
                  ["Languages", portfolio.languages.length],
                ].map(([label, value]) => (
                  <div className="rounded-xl bg-muted p-3" key={label}>
                    <dt className="text-xs text-muted-foreground">{label}</dt>
                    <dd className="mt-1 text-xl font-semibold">{value}</dd>
                  </div>
                ))}
              </dl>
              <ul
                className="mt-4 flex flex-wrap gap-2"
                aria-label="Portfolio languages"
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
                      <Badge variant="success">Observed merged</Badge>
                    </div>
                    <p className="mt-2 text-sm text-muted-foreground">
                      {item.summary}
                    </p>
                    <a
                      className="mt-3 inline-flex rounded-md text-sm font-semibold text-accent outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
                      href={item.url}
                      rel="noreferrer"
                      target="_blank"
                    >
                      View canonical PR
                    </a>
                  </li>
                ))}
              </ul>
              <p className="mt-4 text-xs text-muted-foreground">
                Analyzed {formatDate(portfolio.analyzedAt)}. Titles are observed
                GitHub facts; summaries only restate repository, merge, and
                title metadata.
              </p>
            </>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
