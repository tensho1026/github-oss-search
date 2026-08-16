import {
  ArrowUpRight,
  BookOpen,
  CheckCircle2,
  CircleMinus,
  Clock3,
  GitFork,
  Languages,
  MessageCircle,
  ShieldCheck,
  Star,
  Tag,
  Users,
} from "lucide-react";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { Badge } from "../../../components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { Icon } from "../../../components/ui/icon";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "../../../components/ui/tooltip";
import type { RepositoryDiscoveryItem } from "../../../shared/api/generated";
import { formatCompactNumber, formatDate } from "../../../shared/lib/format";
import { useI18n } from "../../../shared/i18n/i18n-context";
import {
  categoryLabel,
  confidenceLabel,
  difficultyPresentation,
  evidencePresentation,
  readinessPresentation,
} from "../model/repository-presentation";
import { BookmarkAction } from "../../account/components/BookmarkAction";

type RepositoryCardProps = {
  item: RepositoryDiscoveryItem;
  rank: number;
};

type SignalProps = {
  available: boolean;
  label: string;
};

function Signal({ available, label }: SignalProps) {
  return (
    <li className="flex items-center gap-2 text-sm">
      <Icon
        className={available ? "text-success" : "text-muted-foreground"}
        icon={available ? CheckCircle2 : CircleMinus}
      />
      <span>{label}</span>
    </li>
  );
}

export function RepositoryCard({ item, rank }: RepositoryCardProps) {
  const { locale, t } = useI18n();
  const readiness = readinessPresentation(item.readiness);
  const difficulty = difficultyPresentation(item.difficulty);
  const documentation = evidencePresentation(item.documentation.status);
  const japanese = item.documentation.japaneseReadme;
  const japaneseEvidence = evidencePresentation(japanese.status);

  return (
    <Card className="overflow-hidden">
      <CardHeader className="border-b border-border bg-muted/25">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="neutral">#{rank}</Badge>
              <Badge variant="info">{categoryLabel(item.category)}</Badge>
              {item.repository.isFork ? (
                <Badge variant="neutral">{t("repository.fork")}</Badge>
              ) : null}
              {item.repository.isArchived ? (
                <Badge variant="warning">{t("repository.archived")}</Badge>
              ) : null}
            </div>
            <CardTitle className="mt-4 text-2xl">
              <a
                className="inline-flex max-w-full items-center gap-2 rounded-md outline-none hover:text-accent focus-visible:ring-2 focus-visible:ring-ring"
                href={item.repository.url}
                rel="noreferrer"
                target="_blank"
              >
                <span className="truncate">{item.repository.fullName}</span>
                <Icon className="shrink-0" icon={ArrowUpRight} />
              </a>
            </CardTitle>
            <CardDescription className="mt-2 max-w-3xl">
              {item.repository.description || t("repository.noDescription")}
            </CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge variant={readiness.tone}>
              {readiness.label} · {item.readiness.score}/100
            </Badge>
            <Badge variant={difficulty.tone}>
              {t("repository.difficulty", {
                label: difficulty.label,
                level: item.difficulty.level,
              })}
            </Badge>
          </div>
        </div>
      </CardHeader>

      <CardContent className="grid gap-6 p-5 sm:p-6">
        <dl className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <Metric
            icon={Star}
            label={t("repository.stars")}
            value={formatCompactNumber(item.popularity.stars, locale)}
          />
          <Metric
            icon={GitFork}
            label={t("repository.forks")}
            value={formatCompactNumber(item.popularity.forks, locale)}
          />
          <Metric
            icon={Users}
            label={t("repository.watchers")}
            value={formatCompactNumber(item.popularity.watchers, locale)}
          />
          <Metric
            icon={MessageCircle}
            label={t("repository.openIssues")}
            value={formatCompactNumber(item.popularity.openIssues, locale)}
          />
          <Metric
            icon={Clock3}
            label={t("repository.lastPush")}
            value={formatDate(item.activity.pushedAt, locale)}
          />
        </dl>

        <div className="flex flex-wrap gap-2">
          {item.language ? (
            <Badge variant="accent">
              <Icon icon={Languages} />
              {item.language}
            </Badge>
          ) : (
            <Badge variant="neutral">
              {t("repository.languageUnavailable")}
            </Badge>
          )}
          <Badge variant={item.license.spdxId ? "success" : "neutral"}>
            <Icon icon={ShieldCheck} />
            {item.license.spdxId ?? t("repository.licenseUnavailable")}
          </Badge>
          {item.technologies.map((technology) => (
            <Badge key={technology} variant="accent">
              {technology}
            </Badge>
          ))}
        </div>

        <div className="grid gap-5 xl:grid-cols-3">
          <section
            aria-labelledby={`${item.repository.fullName}-readiness`}
            className="rounded-xl border border-border bg-muted/30 p-4"
          >
            <h3
              className="font-semibold"
              id={`${item.repository.fullName}-readiness`}
            >
              {t("repository.readinessReason")}
            </h3>
            <ul className="mt-3 grid gap-2 text-sm leading-6 text-muted-foreground">
              {item.readiness.reasons.map((reason) => (
                <li key={reason}>• {reason}</li>
              ))}
            </ul>
            <p className="mt-4 text-xs text-muted-foreground">
              {t("repository.starterIssues", {
                goodFirst: formatCompactNumber(
                  item.readiness.goodFirstIssues,
                  locale,
                ),
                helpWanted: formatCompactNumber(
                  item.readiness.helpWantedIssues,
                  locale,
                ),
              })}
            </p>
          </section>

          <section
            aria-labelledby={`${item.repository.fullName}-documentation`}
            className="rounded-xl border border-border bg-muted/30 p-4"
          >
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3
                className="font-semibold"
                id={`${item.repository.fullName}-documentation`}
              >
                {t("repository.documents")}
              </h3>
              <Badge variant={documentation.tone}>{documentation.label}</Badge>
            </div>
            <ul className="mt-3 grid gap-2">
              <Signal
                available={item.documentation.readmeAvailable}
                label="README"
              />
              <Signal
                available={item.documentation.contributingGuide}
                label={t("repository.contributingGuide")}
              />
              <Signal
                available={item.documentation.codeOfConduct}
                label={t("repository.codeOfConduct")}
              />
              <Signal
                available={item.documentation.securityPolicy}
                label={t("repository.securityPolicy")}
              />
            </ul>
          </section>

          <section
            aria-labelledby={`${item.repository.fullName}-japanese`}
            className="rounded-xl border border-border bg-muted/30 p-4"
          >
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3
                className="font-semibold"
                id={`${item.repository.fullName}-japanese`}
              >
                {t("repository.japaneseEvidence")}
              </h3>
              <Badge variant={japaneseEvidence.tone}>
                {japaneseEvidence.label}
              </Badge>
            </div>
            <p className="mt-3 text-sm font-medium">
              {japanese.status === "unavailable"
                ? t("repository.notAnalyzed")
                : japanese.detected
                  ? t("repository.japaneseDetected")
                  : t("repository.japaneseNotDetected")}
            </p>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  className="mt-2 rounded-md text-left text-xs text-muted-foreground underline decoration-dotted underline-offset-4 outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  type="button"
                >
                  {confidenceLabel(japanese.confidence)}
                </button>
              </TooltipTrigger>
              <TooltipContent className="max-w-xs">
                {t("repository.japaneseHeuristic")}
              </TooltipContent>
            </Tooltip>
            {japanese.status !== "unavailable" ? (
              <p className="mt-3 text-xs leading-5 text-muted-foreground">
                {t("repository.japaneseCounts", {
                  bytes: formatCompactNumber(japanese.analyzedBytes, locale),
                  japanese: formatCompactNumber(japanese.japaneseRunes, locale),
                  letters: formatCompactNumber(japanese.letterRunes, locale),
                })}
              </p>
            ) : null}
          </section>
        </div>

        <div className="grid gap-5 xl:grid-cols-2">
          <section aria-label={t("repository.difficultyReasons")}>
            <h3 className="font-semibold">
              {t("repository.difficultyEvidence")}
            </h3>
            <ul className="mt-2 grid gap-2 text-sm leading-6 text-muted-foreground">
              {item.difficulty.reasons.map((reason) => (
                <li key={reason}>• {reason}</li>
              ))}
            </ul>
          </section>
          <section aria-label={t("repository.topicsLabel")}>
            <h3 className="font-semibold">{t("repository.topics")}</h3>
            {item.topics.length > 0 ? (
              <ul className="mt-2 flex flex-wrap gap-2">
                {item.topics.map((topic) => (
                  <li key={topic}>
                    <Badge variant="neutral">
                      <Icon icon={Tag} />
                      {topic}
                    </Badge>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-2 text-sm text-muted-foreground">
                {t("repository.noTopics")}
              </p>
            )}
          </section>
        </div>

        {item.warnings.length > 0 ? (
          <Alert variant="warning">
            <AlertTitle>{t("repository.evidencePartial")}</AlertTitle>
            <AlertDescription>
              {item.warnings.map((warning) => warning.message).join(" ")}
            </AlertDescription>
          </Alert>
        ) : null}

        <p className="flex items-center gap-2 text-xs text-muted-foreground">
          <Icon icon={BookOpen} />
          {t("repository.updated", {
            date: formatDate(item.activity.updatedAt, locale),
          })}
        </p>
        <div>
          <BookmarkAction
            request={{
              repositoryName: item.repository.name,
              repositoryOwner: item.repository.owner,
              targetType: "repository",
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function Metric({
  icon,
  label,
  value,
}: {
  icon: typeof Star;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-xl bg-muted p-3">
      <dt className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon icon={icon} />
        {label}
      </dt>
      <dd className="mt-2 text-sm font-semibold">{value}</dd>
    </div>
  );
}
