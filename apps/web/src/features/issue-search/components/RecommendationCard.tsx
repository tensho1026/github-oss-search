import { ArrowUpRight, Clock3, Code2, MessageCircle, Star } from "lucide-react";
import { Link, useLocation } from "react-router";

import { Alert, AlertDescription } from "../../../components/ui/alert";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "../../../components/ui/card";
import { Icon } from "../../../components/ui/icon";
import type { IssueSearchItem } from "../../../shared/api/generated";
import { appRoutes, externalLinks } from "../../../shared/config/app-config";
import {
  formatCompactNumber,
  formatDate,
  formatDuration,
  formatPercentage,
  formatRating,
} from "../../../shared/lib/format";
import { cn } from "../../../shared/lib/cn";
import { issueDetailSearchParameters } from "../../../shared/lib/issue-detail-location";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { BookmarkAction } from "../../account/components/BookmarkAction";
import { IssueClaimAction } from "../../account/components/IssueClaimAction";
import {
  scorePresentation,
  skillPresentation,
  warningPresentation,
} from "../model/search-presentation";

type RecommendationCardProps = {
  item: IssueSearchItem;
  rank: number;
};

export function RecommendationCard({ item, rank }: RecommendationCardProps) {
  const { locale, t } = useI18n();
  const location = useLocation();
  const score = scorePresentation(item.recommendation.score);
  const scoreLabel = t(
    item.recommendation.score >= 75
      ? "recommendation.strong"
      : item.recommendation.score >= 50
        ? "recommendation.promising"
        : "recommendation.careful",
  );
  const detailRoute = appRoutes.issue(
    item.repository.owner,
    item.repository.name,
    item.issue.number,
  );
  const detailParameters = issueDetailSearchParameters(
    `${location.pathname}${location.search}`,
  );
  const detailPath =
    detailParameters.size > 0
      ? `${detailRoute}?${detailParameters.toString()}`
      : detailRoute;
  const maintainerResponse = item.recommendation.maintainerResponse;
  const healthLabels = {
    activity: t("recommendation.healthActivity"),
    beginner_friendly: t("recommendation.healthBeginner"),
    community: t("recommendation.healthCommunity"),
    security: t("recommendation.healthSecurity"),
  } as const;
  return (
    <Card aria-labelledby={`issue-result-${rank}`} className="overflow-hidden">
      <CardHeader className="border-b border-border bg-muted/35">
        <div className="flex flex-wrap items-start justify-between gap-5">
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              <span className="font-mono">#{rank}</span>
              <span>{item.repository.fullName}</span>
              <span aria-hidden="true">·</span>
              <span className="inline-flex items-center gap-1">
                <Icon className="size-3.5" icon={Star} />
                {formatCompactNumber(item.repository.stars, locale)}
              </span>
              <span aria-hidden="true">·</span>
              <span>
                {t("recommendation.updated", {
                  date: formatDate(item.issue.updatedAt, locale),
                })}
              </span>
            </div>
            <h2
              className="mt-3 text-xl leading-7 font-semibold tracking-[-0.03em]"
              id={`issue-result-${rank}`}
            >
              {item.issue.title}
            </h2>
          </div>
          <div
            aria-label={t("recommendation.scoreMeter", {
              label: scoreLabel,
              score: item.recommendation.score,
            })}
            aria-valuemax={100}
            aria-valuemin={0}
            aria-valuenow={item.recommendation.score}
            className={cn(
              "grid size-20 shrink-0 place-items-center rounded-2xl border text-center",
              score.className,
            )}
            role="meter"
          >
            <span>
              <strong className="block font-mono text-2xl leading-none">
                {item.recommendation.score}
              </strong>
              <span className="mt-1 block text-[0.65rem] font-semibold uppercase">
                {scoreLabel}
              </span>
            </span>
          </div>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          {item.repository.mainLanguage ? (
            <Badge variant="info">
              <Icon className="size-3.5" icon={Code2} />
              {item.repository.mainLanguage}
            </Badge>
          ) : null}
          <Badge variant="neutral">
            {t("recommendation.difficulty", {
              label: item.difficulty.label,
              level: item.difficulty.level,
            })}
          </Badge>
          <Badge
            aria-label={t("recommendation.staleStatus", {
              state: item.recommendation.stale.state,
            })}
            variant={staleBadgeVariant(item.recommendation.stale.state)}
          >
            {t("recommendation.staleCheck", {
              state: item.recommendation.stale.state,
            })}
          </Badge>
          <Badge variant="neutral">
            <Icon className="size-3.5" icon={Clock3} />
            {item.effort.label}
          </Badge>
          <Badge variant="neutral">
            <Icon className="size-3.5" icon={MessageCircle} />
            {t("recommendation.comments", {
              count: formatCompactNumber(item.issue.comments, locale),
            })}
          </Badge>
          {item.repository.isArchived ? (
            <Badge variant="warning">{t("recommendation.archived")}</Badge>
          ) : null}
        </div>
        <div
          aria-label={t("recommendation.healthSummary")}
          className="mt-2 flex flex-wrap gap-2"
        >
          {item.healthSummary.map((category) => (
            <Badge key={category.name} variant="neutral">
              {healthLabels[category.name]}{" "}
              {category.score === null ? "?" : category.score}
            </Badge>
          ))}
        </div>
      </CardHeader>

      <CardContent className="grid gap-5 p-5 sm:p-6">
        <section aria-label={t("recommendation.issueLabels")}>
          <p className="text-xs font-semibold tracking-[0.12em] text-muted-foreground uppercase">
            {t("recommendation.labels")}
          </p>
          <div className="mt-2 flex flex-wrap gap-2">
            {item.issue.labels.length > 0 ? (
              item.issue.labels.map((label) => (
                <Badge key={label} variant="accent">
                  {label}
                </Badge>
              ))
            ) : (
              <span className="text-sm text-muted-foreground">
                {t("recommendation.noLabels")}
              </span>
            )}
          </div>
        </section>

        <section
          aria-labelledby={`skill-match-${rank}`}
          className="rounded-xl border border-border bg-muted/35 p-4"
        >
          <div className="flex flex-wrap items-end justify-between gap-3">
            <div>
              <h3 className="font-semibold" id={`skill-match-${rank}`}>
                {t("recommendation.skillMatch")}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("recommendation.skillSummary", {
                  denominator: item.recommendation.skillMatch.denominator,
                  matched: item.recommendation.skillMatch.matched,
                })}
                {item.recommendation.skillMatch.partial > 0
                  ? ` · ${t("recommendation.partialMatch", {
                      count: item.recommendation.skillMatch.partial,
                    })}`
                  : ""}
              </p>
            </div>
            <strong className="font-mono text-xl text-accent">
              {formatPercentage(item.recommendation.skillMatch.percentage)}
            </strong>
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            {item.recommendation.skillMatch.skills.length > 0 ? (
              item.recommendation.skillMatch.skills.map((skill) => (
                <Badge
                  key={`${skill.technology}-${skill.status}`}
                  variant={skillPresentation(skill.status)}
                >
                  {skill.technology}: {skill.status}
                </Badge>
              ))
            ) : (
              <span className="text-sm text-muted-foreground">
                {t("recommendation.noSkillEvidence")}
              </span>
            )}
          </div>
          <p className="mt-3 text-xs text-muted-foreground">
            {item.recommendation.skillMatch.personalized
              ? t("recommendation.personalizedEvidence", {
                  status: item.recommendation.skillMatch.status,
                  version: item.recommendation.skillMatch.version,
                })
              : t("recommendation.explicitEvidence")}
          </p>
        </section>

        <section
          aria-labelledby={`maintainer-response-${rank}`}
          className="rounded-xl border border-border bg-muted/35 p-4"
        >
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 className="font-semibold" id={`maintainer-response-${rank}`}>
                {t("recommendation.maintainerResponse")}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("recommendation.maintainerResponseDescription")}
              </p>
            </div>
            {maintainerResponse.status === "available" ? (
              <div className="text-right">
                <strong
                  aria-label={t("recommendation.rating", {
                    label: maintainerResponse.label,
                    level: maintainerResponse.level,
                  })}
                  className="block tracking-[0.08em] text-accent"
                >
                  {formatRating(maintainerResponse.level)}
                </strong>
                <span className="text-xs font-semibold">
                  {maintainerResponse.label}
                </span>
              </div>
            ) : (
              <Badge variant="neutral">{t("recommendation.unavailable")}</Badge>
            )}
          </div>
          <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
            <div>
              <dt className="text-xs text-muted-foreground">
                {t("recommendation.medianFirstResponse")}
              </dt>
              <dd className="mt-1 font-medium">
                {durationValue(
                  maintainerResponse.firstIssueResponse,
                  locale,
                  t("recommendation.unavailable"),
                )}
              </dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">
                {t("recommendation.prMergeTime")}
              </dt>
              <dd className="mt-1 font-medium">
                {durationValue(
                  maintainerResponse.pullRequestMerge,
                  locale,
                  t("recommendation.unavailable"),
                )}
              </dd>
            </div>
          </dl>
          <p className="mt-3 text-xs leading-5 text-muted-foreground">
            {t("recommendation.maintainerSample", {
              confidence: maintainerResponse.confidence,
              count: formatCompactNumber(maintainerResponse.sampleSize, locale),
            })}
          </p>
        </section>

        <section aria-labelledby={`reasons-${rank}`}>
          <h3 className="font-semibold" id={`reasons-${rank}`}>
            {t("recommendation.reasons")}
          </h3>
          <ul className="mt-2 grid gap-2 text-sm leading-6 text-muted-foreground">
            {item.recommendation.reasons.map((reason) => (
              <li className="flex gap-2" key={reason}>
                <span aria-hidden="true" className="text-accent">
                  •
                </span>
                <span>{reason}</span>
              </li>
            ))}
          </ul>
        </section>

        {item.recommendation.warnings.length > 0 ? (
          <section
            aria-label={t("recommendation.warnings")}
            className="grid gap-2"
          >
            {item.recommendation.warnings.map((warning) => (
              <Alert
                key={`${warning.code}-${warning.message}`}
                variant={warningPresentation(warning.severity)}
              >
                <AlertDescription>{warning.message}</AlertDescription>
              </Alert>
            ))}
          </section>
        ) : null}
      </CardContent>

      <CardFooter className="flex-wrap border-t border-border bg-muted/20 p-5 sm:p-6">
        <Button asChild>
          <Link to={detailPath}>{t("recommendation.details")}</Link>
        </Button>
        <Button asChild variant="outline">
          <a
            href={externalLinks.gitHubIssue(
              item.repository.owner,
              item.repository.name,
              item.issue.number,
            )}
            rel="noreferrer"
            target="_blank"
          >
            {t("recommendation.openGitHub")}
            <Icon icon={ArrowUpRight} />
          </a>
        </Button>
        <BookmarkAction
          request={{
            issueNumber: item.issue.number,
            repositoryName: item.repository.name,
            repositoryOwner: item.repository.owner,
            targetType: "issue",
          }}
        />
        <IssueClaimAction
          request={{
            issueNumber: item.issue.number,
            repositoryName: item.repository.name,
            repositoryOwner: item.repository.owner,
          }}
        />
      </CardFooter>
    </Card>
  );
}

function durationValue(
  metric: {
    medianSeconds: number | null;
    status: string;
  },
  locale: string,
  unavailable: string,
) {
  return metric.status === "available"
    ? formatDuration(metric.medianSeconds, locale)
    : unavailable;
}

function staleBadgeVariant(state: string) {
  switch (state) {
    case "fresh":
      return "success" as const;
    case "stale":
      return "warning" as const;
    default:
      return "neutral" as const;
  }
}
