import { ArrowLeft } from "lucide-react";
import { useEffect, useRef, type ReactNode } from "react";
import { Link } from "react-router";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { Icon } from "../../../components/ui/icon";
import type {
  CountAggregate,
  DurationAggregate,
  Evidence,
  IssueDetailEnvelope,
  RatioAggregate,
} from "../../../shared/api/generated";
import { externalLinks } from "../../../shared/config/app-config";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { cn } from "../../../shared/lib/cn";
import {
  formatCompactNumber,
  formatDate,
  formatDuration,
  formatPercentage,
  formatRating,
} from "../../../shared/lib/format";
import {
  scorePresentation,
  skillPresentation,
  warningPresentation,
} from "../../issue-search/model/search-presentation";
import { BookmarkAction } from "../../account/components/BookmarkAction";
import { IssueClaimAction } from "../../account/components/IssueClaimAction";
import {
  categoryLabel,
  qualitySignalLabel,
  repositorySignalLabel,
  scopeAreaLabel,
  scoreComponentLabel,
  signalPresentation,
} from "../model/detail-presentation";
import { SafeIssueBody } from "./SafeIssueBody";

type Props = {
  envelope: IssueDetailEnvelope;
  returnTo: string;
};

type Sample = Pick<
  CountAggregate,
  "confidence" | "sampleSize" | "truncated" | "windowDays"
>;

function Section({ children, title }: { children: ReactNode; title: string }) {
  return (
    <Card className="min-w-0 overflow-hidden">
      <CardHeader className="border-b border-border bg-muted/25">
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-5 sm:p-6">{children}</CardContent>
    </Card>
  );
}

function EvidenceList({ items }: { items: Evidence[] }) {
  return items.length > 0 ? (
    <ul className="mt-2 grid gap-1 text-xs leading-5 text-muted-foreground">
      {items.map((item) => (
        <li key={`${item.ruleId}-${item.source}-${item.description}`}>
          {item.description}
        </li>
      ))}
    </ul>
  ) : null;
}

function Facts({ items }: { items: Array<[string, ReactNode]> }) {
  return (
    <dl className="grid grid-cols-2 gap-4">
      {items.map(([label, value]) => (
        <div key={label}>
          <dt className="text-xs text-muted-foreground">{label}</dt>
          <dd className="mt-1 font-medium">{value}</dd>
        </div>
      ))}
    </dl>
  );
}

function RepositoryHealthDashboard({
  dashboard,
}: {
  dashboard: IssueDetailEnvelope["data"]["healthDashboard"];
}) {
  const { locale, t } = useI18n();
  const healthCategoryLabels = {
    activity: t("detail.healthActivity"),
    beginner_friendly: t("detail.healthBeginner"),
    community: t("detail.healthCommunity"),
    security: t("detail.healthSecurity"),
  } as const;
  return (
    <Section title={t("detail.healthTitle")}>
      <p className="mb-4 text-sm text-muted-foreground">
        {t("detail.healthDescription", { version: dashboard.scoreVersion })}
      </p>
      <div className="grid gap-3 sm:grid-cols-2">
        {dashboard.categories.map((category) => (
          <details
            className="rounded-xl border border-border bg-muted/25 p-3"
            key={category.name}
          >
            <summary className="cursor-pointer font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
              {healthCategoryLabels[category.name]}{" "}
              {category.score === null
                ? t("detail.unavailable")
                : category.score}
              <Badge className="ml-2" variant="neutral">
                {t("detail.healthStatus", {
                  confidence: category.confidence,
                  status: category.status,
                })}
              </Badge>
            </summary>
            <ul className="mt-3 grid gap-2 text-xs">
              {category.components.map((component) => (
                <li key={component.key}>
                  <span className="font-semibold">
                    {t("detail.healthWeight", {
                      component: component.key.replaceAll("_", " "),
                      weight: component.weight,
                    })}
                  </span>
                  <span className="text-muted-foreground">
                    {" "}
                    ·{" "}
                    {component.score === null
                      ? t("detail.unavailable")
                      : component.score}{" "}
                    · {component.source}
                  </span>
                  <p className="text-muted-foreground">
                    {component.description}
                  </p>
                </li>
              ))}
            </ul>
            {category.warnings.map((warning) => (
              <p className="mt-2 text-xs text-warning" key={warning}>
                {warning}
              </p>
            ))}
            <p className="mt-2 text-xs text-muted-foreground">
              {t("detail.healthAnalyzed", {
                date: formatDate(category.analyzedAt, locale),
              })}
              {category.sourceVersion
                ? ` · ${t("detail.healthUpstream", {
                    version: category.sourceVersion,
                  })}`
                : ""}
            </p>
          </details>
        ))}
      </div>
      <p className="mt-4 text-xs text-muted-foreground">
        {t("detail.healthSecurityNote")}
      </p>
    </Section>
  );
}

function Metric({
  detail,
  label,
  sample,
  value,
}: {
  detail?: string;
  label: string;
  sample: Sample;
  value: string;
}) {
  const { locale, t } = useI18n();
  return (
    <div className="rounded-xl border border-border bg-muted/25 p-3">
      <dt className="text-xs font-semibold text-muted-foreground uppercase">
        {label}
      </dt>
      <dd className="mt-1 font-mono text-xl font-semibold">{value}</dd>
      {detail ? (
        <p className="text-xs text-muted-foreground">{detail}</p>
      ) : null}
      <p className="mt-1 text-xs leading-5 text-muted-foreground">
        {t("detail.sampleSummary", {
          bounded: sample.truncated ? t("detail.bounded") : "",
          confidence: sample.confidence,
          count: formatCompactNumber(sample.sampleSize, locale),
          days: formatCompactNumber(sample.windowDays, locale),
        })}
      </p>
    </div>
  );
}

function countValue(
  metric: CountAggregate,
  unavailable: string,
  locale: string,
) {
  return metric.status === "available" && metric.value !== null
    ? formatCompactNumber(metric.value, locale)
    : unavailable;
}

function durationValue(
  metric: DurationAggregate,
  unavailable: string,
  locale: string,
) {
  return metric.status === "available"
    ? formatDuration(metric.medianSeconds, locale)
    : unavailable;
}

function ratioValue(metric: RatioAggregate, unavailable: string) {
  return metric.status === "available" && metric.percentage !== null
    ? formatPercentage(metric.percentage)
    : unavailable;
}

export function IssueDetailDashboard({ envelope, returnTo }: Props) {
  const { locale, t } = useI18n();
  const { data, meta } = envelope;
  const titleRef = useRef<HTMLHeadingElement>(null);
  const score = scorePresentation(data.recommendation.score);
  const activityMetrics: Array<{
    detail?: string;
    label: string;
    sample: Sample;
    value: string;
  }> = [
    {
      label: t("detail.contributors"),
      sample: data.activity.contributors,
      value: countValue(
        data.activity.contributors,
        t("detail.unavailable"),
        locale,
      ),
    },
    {
      label: t("detail.prOpened"),
      sample: data.activity.pullRequestsOpened,
      value: countValue(
        data.activity.pullRequestsOpened,
        t("detail.unavailable"),
        locale,
      ),
    },
    {
      label: t("detail.stalePr"),
      sample: data.activity.staleOpenPullRequests,
      value: countValue(
        data.activity.staleOpenPullRequests,
        t("detail.unavailable"),
        locale,
      ),
    },
    {
      label: t("detail.unansweredIssues"),
      sample: data.activity.unansweredIssues,
      value: countValue(
        data.activity.unansweredIssues,
        t("detail.unavailable"),
        locale,
      ),
    },
    {
      detail: t("detail.openedSummary", {
        denominator: data.activity.pullRequestMerge.denominator ?? "—",
        numerator: data.activity.pullRequestMerge.numerator ?? "—",
      }),
      label: t("detail.mergeRate"),
      sample: data.activity.pullRequestMerge,
      value: ratioValue(
        data.activity.pullRequestMerge,
        t("detail.unavailable"),
      ),
    },
    ...(
      [
        [t("detail.firstIssueResponse"), data.activity.issueResponse],
        [t("detail.firstPrReview"), data.activity.pullRequestReview],
        [t("detail.prMergeTime"), data.activity.pullRequestMergeTime],
      ] as Array<[string, DurationAggregate]>
    ).map(([label, metric]) => ({
      detail: t("detail.percentile90", {
        value:
          metric.status === "available"
            ? formatDuration(metric.percentile90Seconds, locale)
            : t("detail.unavailable"),
      }),
      label,
      sample: metric,
      value: durationValue(metric, t("detail.unavailable"), locale),
    })),
  ];

  useEffect(() => {
    titleRef.current?.focus({ preventScroll: true });
  }, []);

  return (
    <>
      <Link
        className="inline-flex min-h-11 items-center gap-2 rounded-lg text-sm font-semibold text-muted-foreground outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        to={returnTo}
      >
        <Icon icon={ArrowLeft} />
        {t("detail.backResults")}
      </Link>

      {data.inspection.incomplete ? (
        <Alert className="mt-4" variant="warning">
          <AlertTitle>{t("detail.partialTitle")}</AlertTitle>
          <AlertDescription>{t("detail.partialDescription")}</AlertDescription>
        </Alert>
      ) : null}

      <Card className="mt-4 overflow-hidden">
        <CardHeader className="border-b border-border bg-muted/30 p-6 sm:p-8">
          <div className="flex flex-wrap items-start justify-between gap-6">
            <div className="min-w-0 flex-1">
              <p className="text-xs text-muted-foreground">
                <a
                  className="rounded font-mono outline-none hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
                  href={externalLinks.gitHubRepository(
                    data.repository.owner,
                    data.repository.name,
                  )}
                  rel="noreferrer"
                  target="_blank"
                >
                  {data.repository.fullName}
                </a>{" "}
                · Issue #{data.issue.number} · {data.issue.state}
              </p>
              <h1
                className="mt-4 max-w-4xl rounded text-3xl font-semibold tracking-[-0.05em] outline-none sm:text-5xl"
                ref={titleRef}
                tabIndex={-1}
              >
                {data.issue.title}
              </h1>
              <p className="mt-4 max-w-3xl leading-7 text-muted-foreground">
                {data.repository.description ||
                  t("detail.noRepositoryDescription")}
              </p>
            </div>
            <div
              aria-label={t("detail.scoreMeter", {
                label: score.label,
                score: data.recommendation.score,
              })}
              aria-valuemax={100}
              aria-valuemin={0}
              aria-valuenow={data.recommendation.score}
              className={cn(
                "grid size-24 shrink-0 place-items-center rounded-2xl border text-center",
                score.className,
              )}
              role="meter"
            >
              <span>
                <strong className="block font-mono text-3xl">
                  {data.recommendation.score}
                </strong>
                <small className="font-semibold uppercase">{score.label}</small>
              </span>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-2">
            <Badge variant="info">
              {data.repository.mainLanguage || t("detail.languageUnknown")}
            </Badge>
            <Badge variant="neutral">
              {t("detail.stars", {
                count: formatCompactNumber(data.repository.stars, locale),
              })}
            </Badge>
            <Badge variant="neutral">
              {t("detail.difficulty", {
                label: data.analysis.difficulty.label,
                level: data.analysis.difficulty.level,
              })}
            </Badge>
            <Badge variant="neutral">{data.analysis.effort.label}</Badge>
            {data.issue.locked ? (
              <Badge variant="warning">{t("detail.locked")}</Badge>
            ) : null}
            {data.repository.isArchived ? (
              <Badge variant="warning">{t("detail.archived")}</Badge>
            ) : null}
          </div>
          <div className="mt-5 flex flex-wrap gap-3">
            <BookmarkAction
              request={{
                issueNumber: data.issue.number,
                repositoryName: data.repository.name,
                repositoryOwner: data.repository.owner,
                targetType: "issue",
              }}
            />
            <IssueClaimAction
              request={{
                issueNumber: data.issue.number,
                repositoryName: data.repository.name,
                repositoryOwner: data.repository.owner,
              }}
            />
          </div>
          <div className="mt-4 flex flex-wrap gap-3">
            <Button asChild>
              <a
                href={externalLinks.gitHubIssue(
                  data.repository.owner,
                  data.repository.name,
                  data.issue.number,
                )}
                rel="noreferrer"
                target="_blank"
              >
                {t("detail.openOriginal")}
              </a>
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-5 sm:p-6">
          <Facts
            items={[
              [
                t("detail.author"),
                <a
                  href={externalLinks.gitHubProfile(data.issue.author.login)}
                  key="author"
                  rel="noreferrer"
                  target="_blank"
                >
                  @{data.issue.author.login} · {data.issue.author.type}
                </a>,
              ],
              [
                t("detail.conversation"),
                t("detail.comments", {
                  count: formatCompactNumber(data.issue.comments, locale),
                }),
              ],
              [
                t("detail.assignees"),
                data.issue.assignees.length > 0
                  ? data.issue.assignees
                      .map((assignee) => `@${assignee}`)
                      .join(", ")
                  : t("detail.unassigned"),
              ],
              [
                t("detail.dates"),
                `${formatDate(data.issue.createdAt, locale)} → ${formatDate(
                  data.issue.updatedAt,
                  locale,
                )}`,
              ],
            ]}
          />
          <div className="mt-4 flex flex-wrap gap-2">
            {data.issue.labels.length > 0 ? (
              data.issue.labels.map((label) => (
                <Badge key={label} variant="accent">
                  {label}
                </Badge>
              ))
            ) : (
              <span className="text-sm text-muted-foreground">
                {t("detail.noLabels")}
              </span>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="mt-6 grid items-start gap-6 lg:grid-cols-[minmax(0,1.28fr)_minmax(19rem,0.72fr)]">
        <div className="grid min-w-0 gap-6">
          <Section title={t("detail.description")}>
            <SafeIssueBody body={data.issue.body} />
          </Section>

          <Section title={t("detail.involves")}>
            <Facts
              items={[
                [
                  t("detail.category"),
                  data.analysis.category.matches.map(categoryLabel).join(", "),
                ],
                [
                  t("repositoryForm.maximumDifficulty"),
                  `${data.analysis.difficulty.label} · ${data.analysis.difficulty.confidence}`,
                ],
                [
                  t("detail.effort"),
                  `${data.analysis.effort.label} · ${data.analysis.effort.confidence}`,
                ],
                [
                  t("detail.scope"),
                  `${data.analysis.scope.fileCount.label} · ${data.analysis.scope.confidence}`,
                ],
              ]}
            />
            <div className="mt-4 flex flex-wrap gap-2">
              {data.analysis.scope.areas.map((area) => (
                <Badge key={area} variant="info">
                  {scopeAreaLabel(area)}
                </Badge>
              ))}
              <Badge
                variant={
                  signalPresentation(data.analysis.scope.databaseChange).tone
                }
              >
                {t("detail.database", {
                  label: signalPresentation(data.analysis.scope.databaseChange)
                    .label,
                })}
              </Badge>
              <Badge variant="neutral">
                {t("detail.analysisConfidence", {
                  confidence: data.analysis.confidence,
                })}
              </Badge>
            </div>
            <EvidenceList items={data.analysis.scope.evidence} />

            <h3 className="mt-6 font-semibold">
              {t("detail.requiredTechnologies")}
            </h3>
            <ul className="mt-3 grid gap-3 sm:grid-cols-2">
              {data.analysis.requiredTechnologies.map((technology) => (
                <li
                  className="rounded-xl border border-border bg-muted/25 p-3"
                  key={`${technology.kind}-${technology.name}`}
                >
                  <strong>{technology.name}</strong>{" "}
                  <Badge variant="neutral">{technology.kind}</Badge>
                  <p className="text-xs text-muted-foreground">
                    {technology.confidence} confidence
                  </p>
                  <EvidenceList items={technology.evidence} />
                </li>
              ))}
            </ul>

            <h3 className="mt-6 font-semibold">
              {t("detail.quality")} · {data.analysis.quality.score}/100 ·{" "}
              {data.analysis.quality.confidence} confidence
            </h3>
            <ul className="mt-3 grid gap-2 sm:grid-cols-2">
              {data.analysis.quality.signals.map((signal) => {
                const state = signalPresentation(signal.state);
                return (
                  <li
                    className="rounded-xl border border-border bg-muted/25 p-3"
                    key={signal.key}
                  >
                    <span className="text-sm font-medium">
                      {qualitySignalLabel(signal.key)}
                    </span>{" "}
                    <Badge variant={state.tone}>{state.label}</Badge>
                    <EvidenceList items={signal.evidence} />
                  </li>
                );
              })}
            </ul>
            <h3 className="mt-6 font-semibold">
              {t("detail.assessmentEvidence")}
            </h3>
            <EvidenceList
              items={[
                ...data.analysis.category.evidence,
                ...data.analysis.difficulty.evidence,
                ...data.analysis.effort.evidence,
              ]}
            />
          </Section>

          <Section title={t("detail.why")}>
            <p className="text-xl font-semibold">
              {t("detail.skillMatch")}: {data.recommendation.skillMatch.matched}
              /{data.recommendation.skillMatch.denominator} ·{" "}
              {formatPercentage(data.recommendation.skillMatch.percentage)}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("detail.partialEvidence", {
                count: data.recommendation.skillMatch.partial,
                status: data.recommendation.skillMatch.status,
                version: data.recommendation.skillMatch.version,
              })}
            </p>
            <div className="mt-3 flex flex-wrap gap-2">
              {data.recommendation.skillMatch.skills.length > 0
                ? data.recommendation.skillMatch.skills.map((skill) => (
                    <Badge
                      key={`${skill.technology}-${skill.status}`}
                      variant={skillPresentation(skill.status)}
                    >
                      {skill.technology}: {skill.status}
                    </Badge>
                  ))
                : t("detail.noSkillEvidence")}
            </div>
            <ul className="mt-4 list-disc space-y-1 pl-5 text-sm">
              {data.recommendation.reasons.map((reason) => (
                <li key={reason}>{reason}</li>
              ))}
            </ul>

            <h3 className="mt-6 font-semibold">{t("detail.breakdown")}</h3>
            <ul className="mt-3 grid gap-3">
              {data.recommendation.breakdown.map((component) => (
                <li
                  className="rounded-xl border border-border bg-muted/25 p-3"
                  key={component.name}
                >
                  <strong>
                    {scoreComponentLabel(component.name)} · {component.score}/
                    {component.maximum}
                  </strong>
                  <progress
                    aria-label={`${scoreComponentLabel(component.name)} score`}
                    className="mt-2 block h-2 w-full accent-accent"
                    max={component.maximum}
                    value={component.score}
                  />
                  <ul className="mt-2 text-xs text-muted-foreground">
                    {component.reasons.map((reason) => (
                      <li key={reason}>{reason}</li>
                    ))}
                  </ul>
                </li>
              ))}
            </ul>

            <div className="mt-5">
              <Badge
                variant={
                  data.recommendation.claim.claimed ? "warning" : "success"
                }
              >
                {t("detail.claimConfidence", {
                  confidence: data.recommendation.claim.confidence,
                  label: data.recommendation.claim.claimed
                    ? t("detail.claimed")
                    : t("detail.notClaimed"),
                })}
              </Badge>
              <EvidenceList items={data.recommendation.claim.evidence} />
            </div>
            <div className="mt-5 rounded-xl border border-border bg-muted/25 p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h3 className="font-semibold">{t("detail.staleTitle")}</h3>
                <Badge
                  variant={
                    data.recommendation.stale.state === "fresh"
                      ? "success"
                      : data.recommendation.stale.state === "stale"
                        ? "warning"
                        : "neutral"
                  }
                >
                  {t("detail.staleStatus", {
                    confidence: data.recommendation.stale.confidence,
                    state: data.recommendation.stale.state,
                  })}
                </Badge>
              </div>
              <p className="mt-2 text-xs leading-5 text-muted-foreground">
                {t("detail.stalePolicy", {
                  bounded: data.recommendation.stale.truncated
                    ? t("detail.bounded")
                    : "",
                  count: formatCompactNumber(
                    data.recommendation.stale.sampleSize,
                    locale,
                  ),
                  freshDays: data.recommendation.stale.freshWithinDays,
                  staleDays: data.recommendation.stale.staleAfterDays,
                  version: data.recommendation.stale.policyVersion,
                })}
              </p>
              <Facts
                items={[
                  [
                    t("detail.lastMeaningfulActivity"),
                    data.recommendation.stale.lastMeaningfulIssueActivityAt
                      ? formatDate(
                          data.recommendation.stale
                            .lastMeaningfulIssueActivityAt,
                          locale,
                        )
                      : t("detail.unknown"),
                  ],
                  [
                    t("detail.lastMaintainerActivity"),
                    data.recommendation.stale.lastMaintainerActivityAt
                      ? formatDate(
                          data.recommendation.stale.lastMaintainerActivityAt,
                          locale,
                        )
                      : t("detail.unknown"),
                  ],
                  [
                    t("detail.lastLinkedPr"),
                    data.recommendation.stale.lastLinkedPullRequestAt
                      ? formatDate(
                          data.recommendation.stale.lastLinkedPullRequestAt,
                          locale,
                        )
                      : t("detail.unknown"),
                  ],
                  [
                    t("detail.analyzed"),
                    formatDate(data.recommendation.stale.analyzedAt, locale),
                  ],
                ]}
              />
              <EvidenceList items={data.recommendation.stale.evidence} />
            </div>
            <div className="mt-4 grid gap-2">
              {data.recommendation.warnings.map((warning) => (
                <Alert
                  key={`${warning.code}-${warning.message}`}
                  variant={warningPresentation(warning.severity)}
                >
                  <AlertTitle>{warning.code.replaceAll("_", " ")}</AlertTitle>
                  <AlertDescription>{warning.message}</AlertDescription>
                  <EvidenceList items={warning.evidence} />
                </Alert>
              ))}
            </div>
          </Section>
        </div>

        <aside
          aria-label={t("detail.repositorySignals")}
          className="grid gap-6"
        >
          <RepositoryHealthDashboard dashboard={data.healthDashboard} />
          <Section title={t("detail.repositorySnapshot")}>
            <Facts
              items={[
                [
                  t("repository.stars"),
                  formatCompactNumber(data.repository.stars, locale),
                ],
                [
                  t("detail.forks"),
                  formatCompactNumber(data.repository.forks, locale),
                ],
                [
                  t("detail.openIssues"),
                  formatCompactNumber(data.repository.openIssues, locale),
                ],
                [t("detail.defaultBranch"), data.repository.defaultBranch],
                [t("detail.repositoryId"), data.repository.id],
                [
                  t("detail.kind"),
                  data.repository.isFork
                    ? t("detail.forkedRepository")
                    : t("detail.sourceRepository"),
                ],
                [
                  t("detail.updated"),
                  formatDate(data.repository.updatedAt, locale),
                ],
                [
                  t("detail.lastPushed"),
                  data.repository.pushedAt
                    ? formatDate(data.repository.pushedAt, locale)
                    : t("detail.unknown"),
                ],
              ]}
            />
          </Section>

          <Section title={t("detail.contributorReadiness")}>
            <ul className="grid gap-2">
              {data.repositoryHealth.map((signal) => {
                const state = signalPresentation(signal.state);
                return (
                  <li
                    className="rounded-xl border border-border bg-muted/25 p-3"
                    key={signal.key}
                  >
                    {repositorySignalLabel(signal.key)}{" "}
                    <Badge variant={state.tone}>{state.label}</Badge>
                    <EvidenceList items={signal.evidence} />
                  </li>
                );
              })}
            </ul>
          </Section>

          <Section title={t("detail.maintainerActivity")}>
            <div className="mb-4 rounded-xl border border-border bg-muted/25 p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="font-semibold">
                    {t("detail.maintainerResponseScore")}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t("detail.maintainerResponseDescription")}
                  </p>
                </div>
                {data.recommendation.maintainerResponse.status ===
                "available" ? (
                  <div className="text-right">
                    <strong
                      aria-label={t("recommendation.rating", {
                        label: data.recommendation.maintainerResponse.label,
                        level: data.recommendation.maintainerResponse.level,
                      })}
                      className="block tracking-[0.08em] text-accent"
                    >
                      {formatRating(
                        data.recommendation.maintainerResponse.level,
                      )}
                    </strong>
                    <span className="text-xs font-semibold">
                      {data.recommendation.maintainerResponse.label}
                    </span>
                  </div>
                ) : (
                  <Badge variant="neutral">{t("detail.unavailable")}</Badge>
                )}
              </div>
              <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                <div>
                  <dt className="text-xs text-muted-foreground">
                    {t("detail.responseCoverage")}
                  </dt>
                  <dd className="mt-1 font-medium">
                    {ratioValue(
                      data.recommendation.maintainerResponse.responseCoverage,
                      t("detail.unavailable"),
                    )}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">
                    {t("detail.responseSample")}
                  </dt>
                  <dd className="mt-1 font-medium">
                    {formatCompactNumber(
                      data.recommendation.maintainerResponse.sampleSize,
                      locale,
                    )}{" "}
                    · {data.recommendation.maintainerResponse.confidence}
                  </dd>
                </div>
              </dl>
              <p className="mt-3 text-xs leading-5 text-muted-foreground">
                {t("detail.maintainerCaveat")}
              </p>
            </div>
            <p className="mb-3 text-sm text-muted-foreground">
              {t("detail.activitySummary", {
                ci: data.activity.ci,
                date: data.activity.lastMeaningfulUpdate
                  ? formatDate(data.activity.lastMeaningfulUpdate, locale)
                  : t("detail.unknown"),
              })}
            </p>
            <dl className="grid gap-2">
              {activityMetrics.map((metric) => (
                <Metric key={metric.label} {...metric} />
              ))}
            </dl>
          </Section>

          <Card>
            <CardContent className="grid gap-1 p-5 text-xs text-muted-foreground">
              <p>
                {t("detail.generated", {
                  date: formatDate(meta.timestamp, locale),
                })}
              </p>
              <p>{t("detail.requestId", { requestId: meta.requestId })}</p>
              <p>
                {t("detail.rateRemaining", {
                  count:
                    meta.rateLimitRemaining === undefined
                      ? t("detail.unavailable")
                      : formatCompactNumber(meta.rateLimitRemaining, locale),
                })}
              </p>
            </CardContent>
          </Card>
        </aside>
      </div>
    </>
  );
}
