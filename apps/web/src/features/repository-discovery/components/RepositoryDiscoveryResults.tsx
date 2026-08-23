import { CircleOff, Gauge, SearchCheck } from "lucide-react";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { Button } from "../../../components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { Icon } from "../../../components/ui/icon";
import { Pagination } from "../../../components/ui/pagination";
import type { RepositoryDiscoveryEnvelope } from "../../../shared/api/generated";
import { formatCompactNumber } from "../../../shared/lib/format";
import { RepositoryCard } from "./RepositoryCard";
import { useI18n } from "../../../shared/i18n/i18n-context";

type RepositoryDiscoveryResultsProps = {
  envelope: RepositoryDiscoveryEnvelope;
  isFetching: boolean;
  relaxed?: boolean;
  onPageChange: (page: number) => void;
};

export function RepositoryDiscoveryResults({
  envelope,
  isFetching,
  relaxed,
  onPageChange,
}: RepositoryDiscoveryResultsProps) {
  const { locale, t } = useI18n();
  const { items, pagination, searchSummary, warnings } = envelope.data;
  if (items.length === 0) {
    const outOfRange =
      pagination.total > 0 &&
      pagination.totalPages > 0 &&
      pagination.page > pagination.totalPages;
    return (
      <Card>
        <CardContent className="grid justify-items-center gap-4 p-8 text-center sm:p-12">
          <span className="grid size-14 place-items-center rounded-2xl bg-muted text-muted-foreground">
            <Icon className="size-6" icon={CircleOff} />
          </span>
          <div>
            <h2 className="text-xl font-semibold">
              {outOfRange
                ? t("repository.noPageTitle")
                : t("repository.emptyTitle")}
            </h2>
            <p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-muted-foreground">
              {outOfRange
                ? t("repository.noPageDescription")
                : t("repository.emptyDescription")}
            </p>
          </div>
          {outOfRange ? (
            <Button onClick={() => onPageChange(1)}>
              {t("issueSearch.returnFirst")}
            </Button>
          ) : (
            <a
              className="rounded-lg text-sm font-semibold text-accent outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
              href="#repository-filters"
            >
              {t("issueSearch.broaden")}
            </a>
          )}
        </CardContent>
      </Card>
    );
  }

  const hasPartialEvidence =
    warnings.length > 0 ||
    searchSummary.enrichmentFailed > 0 ||
    searchSummary.enrichmentIncomplete ||
    searchSummary.githubIncomplete;
  const firstRank = (pagination.page - 1) * pagination.perPage + 1;

  return (
    <section
      aria-labelledby="repository-results-heading"
      className="grid gap-5"
    >
      <Card>
        <CardHeader className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
          <div className="min-w-0">
            <p className="font-mono text-xs tracking-[0.14em] text-accent uppercase">
              {t("repository.orderedEyebrow")}
            </p>
            <CardTitle
              className="mt-2 text-2xl"
              id="repository-results-heading"
            >
              {t("repository.eligibleCount", {
                count: formatCompactNumber(pagination.total, locale),
              })}
            </CardTitle>
          </div>
          <span className="inline-flex shrink-0 items-center gap-2 self-start whitespace-nowrap rounded-full border border-border bg-muted px-3 py-2 text-xs text-muted-foreground sm:self-center">
            <Icon icon={Gauge} />
            {t("issueSearch.checkedSummary", {
              checked: formatCompactNumber(
                searchSummary.candidatesChecked,
                locale,
              ),
              enriched: formatCompactNumber(
                searchSummary.enrichmentAttempted,
                locale,
              ),
            })}
          </span>
        </CardHeader>
        <CardContent className="grid gap-2 text-xs text-muted-foreground sm:grid-cols-2">
          <p>
            {t("repository.upstreamTotal", {
              count: formatCompactNumber(searchSummary.upstreamTotal, locale),
            })}
          </p>
          {envelope.meta.rateLimitRemaining !== undefined ? (
            <p className="sm:text-right">
              {t("issueSearch.rateRemaining", {
                count: formatCompactNumber(
                  envelope.meta.rateLimitRemaining,
                  locale,
                ),
              })}
            </p>
          ) : null}
        </CardContent>
      </Card>

      {relaxed ? (
        <Alert variant="warning">
          <AlertTitle>{t("search.relaxedTitle")}</AlertTitle>
        </Alert>
      ) : null}

      {hasPartialEvidence ? (
        <Alert variant="warning">
          <AlertTitle>{t("repository.partialTitle")}</AlertTitle>
          <AlertDescription>
            {t("repository.partialDescription")}{" "}
            {warnings.map((warning) => warning.message).join(" ")}
          </AlertDescription>
        </Alert>
      ) : (
        <Alert variant="success">
          <AlertTitle>{t("issueSearch.completeTitle")}</AlertTitle>
          <AlertDescription>
            <span className="inline-flex items-center gap-2">
              <Icon icon={SearchCheck} />
              {t("repository.completeDescription")}
            </span>
          </AlertDescription>
        </Alert>
      )}

      <ol className="grid gap-5">
        {items.map((item, index) => (
          <li key={item.repository.fullName}>
            <RepositoryCard item={item} rank={firstRank + index} />
          </li>
        ))}
      </ol>

      <Card>
        <CardContent className="p-5 sm:p-6">
          <Pagination
            ariaLabel={t("repository.pagination")}
            disabled={isFetching}
            hasNext={pagination.hasNext}
            onPageChange={onPageChange}
            page={pagination.page}
            totalPages={pagination.totalPages}
          />
        </CardContent>
      </Card>
    </section>
  );
}
