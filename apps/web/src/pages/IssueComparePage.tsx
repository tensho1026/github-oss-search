import { ArrowLeft, ExternalLink } from "lucide-react";
import { useMemo } from "react";
import { Link, useLocation } from "react-router";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Icon } from "../components/ui/icon";
import { useIssueComparison } from "../features/issue-compare/api/useIssueComparison";
import { decodeCompareLocation } from "../features/issue-compare/model/compare-location";
import type { IssueDetail } from "../shared/api/generated";
import { appRoutes, externalLinks } from "../shared/config/app-config";
import { formatDuration, formatPercentage } from "../shared/lib/format";
import { useI18n } from "../shared/i18n/i18n-context";

export function IssueComparePage() {
  const { locale, t } = useI18n();
  const routeLocation = useLocation();
  const location = useMemo(
    () => decodeCompareLocation(new URLSearchParams(routeLocation.search)),
    [routeLocation.search],
  );
  const queries = useIssueComparison(location);

  if (!location.valid) {
    return (
      <CompareMessage
        description={location.message}
        title={t("compare.invalid")}
      />
    );
  }
  if (queries.some((query) => query.isPending)) {
    return <CompareMessage title={t("compare.loading")} />;
  }
  const failed = queries.filter((query) => query.error);
  if (failed.length > 0) {
    return (
      <CompareMessage
        description={t("compare.errorDescription", { count: failed.length })}
        title={t("compare.error")}
      />
    );
  }
  const issues = queries.flatMap((query) =>
    query.data ? [query.data.data] : [],
  );

  return (
    <div className="mx-auto min-h-[68vh] w-full max-w-7xl px-5 py-10 sm:px-8 lg:px-10">
      <Button asChild size="small" variant="ghost">
        <Link to={location.returnTo}>
          <Icon icon={ArrowLeft} />
          {t("compare.back")}
        </Link>
      </Button>
      <header className="mt-5 max-w-3xl">
        <Badge variant="accent">{t("compare.badge")}</Badge>
        <h1 className="mt-4 text-4xl font-semibold tracking-[-0.055em]">
          {t("compare.title")}
        </h1>
        <p className="mt-4 text-muted-foreground">{t("compare.description")}</p>
      </header>
      <div className="mt-8 overflow-x-auto pb-4">
        <div
          className="grid min-w-[52rem] gap-4"
          style={{
            gridTemplateColumns: `repeat(${issues.length}, minmax(16rem, 1fr))`,
          }}
        >
          {issues.map((issue) => (
            <ComparisonColumn
              issue={issue}
              key={`${issue.repository.owner}/${issue.repository.name}#${issue.issue.number}`}
              locale={locale}
            />
          ))}
        </div>
      </div>
      <Alert variant="warning">
        <AlertTitle>{t("compare.evidenceTitle")}</AlertTitle>
        <AlertDescription>{t("compare.evidenceDescription")}</AlertDescription>
      </Alert>
    </div>
  );
}

function ComparisonColumn({
  issue,
  locale,
}: {
  issue: IssueDetail;
  locale: string;
}) {
  const { t } = useI18n();
  const response = issue.recommendation.maintainerResponse;
  return (
    <Card className="overflow-hidden">
      <CardHeader className="border-b border-border bg-muted/35">
        <p className="font-mono text-xs text-muted-foreground">
          {issue.repository.owner}/{issue.repository.name}#{issue.issue.number}
        </p>
        <CardTitle className="mt-2 text-xl">{issue.issue.title}</CardTitle>
        <strong className="mt-3 font-mono text-3xl text-accent">
          {issue.recommendation.score}
          <span className="text-sm">/100</span>
        </strong>
      </CardHeader>
      <CardContent className="grid gap-5 p-5">
        <dl className="grid gap-4">
          <Metric
            label={t("compare.skillMatch")}
            value={formatPercentage(issue.recommendation.skillMatch.percentage)}
          />
          <Metric
            label={t("compare.difficulty")}
            value={`${issue.analysis.difficulty.label} (${issue.analysis.difficulty.level}/5)`}
          />
          <Metric
            label={t("compare.effort")}
            value={issue.analysis.effort.label}
          />
          <Metric
            label={t("compare.stale")}
            value={issue.recommendation.stale.state}
          />
          <Metric
            label={t("compare.claim")}
            value={
              issue.recommendation.claim.claimed
                ? t("compare.claimed")
                : t("compare.notClaimed")
            }
          />
          <Metric
            label={t("compare.response")}
            value={
              response.status === "available"
                ? formatDuration(
                    response.firstIssueResponse.medianSeconds,
                    locale,
                  )
                : t("recommendation.unavailable")
            }
          />
        </dl>
        <section>
          <h2 className="text-sm font-semibold">{t("compare.health")}</h2>
          <div className="mt-2 flex flex-wrap gap-2">
            {issue.healthDashboard.categories.map((category) => (
              <Badge key={category.name} variant="neutral">
                {category.name}: {category.score ?? "?"}
              </Badge>
            ))}
          </div>
        </section>
        <section>
          <h2 className="text-sm font-semibold">
            {t("recommendation.warnings")}
          </h2>
          <p className="mt-2 text-sm text-muted-foreground">
            {issue.recommendation.warnings.length > 0
              ? issue.recommendation.warnings
                  .map((warning) => warning.message)
                  .join(" ")
              : t("compare.noWarnings")}
          </p>
        </section>
        <div className="flex flex-wrap gap-2">
          <Button asChild size="small">
            <Link
              to={appRoutes.issue(
                issue.repository.owner,
                issue.repository.name,
                issue.issue.number,
              )}
            >
              {t("recommendation.details")}
            </Link>
          </Button>
          <Button asChild size="small" variant="outline">
            <a
              href={externalLinks.gitHubIssue(
                issue.repository.owner,
                issue.repository.name,
                issue.issue.number,
              )}
              rel="noreferrer"
              target="_blank"
            >
              GitHub <Icon icon={ExternalLink} />
            </a>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-semibold">{value}</dd>
    </div>
  );
}

function CompareMessage({
  title,
  description,
}: {
  title: string;
  description?: string;
}) {
  const { t } = useI18n();
  return (
    <div className="mx-auto grid min-h-[68vh] w-full max-w-3xl content-center gap-5 px-5">
      <Alert variant="warning">
        <AlertTitle>{title}</AlertTitle>
        {description ? (
          <AlertDescription>{description}</AlertDescription>
        ) : null}
      </Alert>
      <Button asChild variant="outline">
        <Link to={appRoutes.search}>{t("compare.back")}</Link>
      </Button>
    </div>
  );
}
