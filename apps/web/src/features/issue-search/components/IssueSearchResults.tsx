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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import type { IssueSearchEnvelope } from "../../../shared/api/generated";
import { formatCompactNumber } from "../../../shared/lib/format";
import { RecommendationCard } from "./RecommendationCard";
import { searchFilterOptions, type IssueSort } from "../model/search-filters";
import { useI18n } from "../../../shared/i18n/i18n-context";

type IssueSearchResultsProps = {
  envelope: IssueSearchEnvelope;
  isFetching: boolean;
  relaxed?: boolean;
  onPageChange: (page: number) => void;
  onSortChange?: (sort: IssueSort) => void;
  sortBy?: IssueSort;
};

export function IssueSearchResults({
  envelope,
  isFetching,
  relaxed,
  onPageChange,
  onSortChange,
  sortBy = "recommendation",
}: IssueSearchResultsProps) {
  const { contributionProfile, items, pagination, searchSummary, warnings } =
    envelope.data;
  const { locale, t } = useI18n();
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
                ? t("issueSearch.noPageTitle")
                : t("issueSearch.emptyTitle")}
            </h2>
            <p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-muted-foreground">
              {outOfRange
                ? t("issueSearch.noPageDescription")
                : t("issueSearch.emptyDescription")}
            </p>
          </div>
          {outOfRange ? (
            <Button onClick={() => onPageChange(1)}>
              {t("issueSearch.returnFirst")}
            </Button>
          ) : (
            <a
              className="rounded-lg text-sm font-semibold text-accent outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring"
              href="#search-filters"
            >
              {t("issueSearch.broaden")}
            </a>
          )}
        </CardContent>
      </Card>
    );
  }

  const hasPartialEvidence =
    warnings.length > 0 || searchSummary.enrichmentFailed > 0;
  const firstRank = (pagination.page - 1) * pagination.perPage + 1;
  return (
    <section aria-labelledby="issue-results-heading" className="grid gap-5">
      <Card>
        <CardHeader className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
          <div className="min-w-0">
            <p className="font-mono text-xs tracking-[0.14em] text-accent uppercase">
              {t("issueSearch.rankedEyebrow")}
            </p>
            <CardTitle className="mt-2 text-2xl" id="issue-results-heading">
              {t("issueSearch.eligibleCount", {
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
          <span className="text-xs text-muted-foreground sm:col-span-2">
            {t("issueSearch.contributionProfile", {
              status: contributionProfile.status,
              version: contributionProfile.version,
            })}
          </span>
        </CardHeader>
        {envelope.meta.rateLimitRemaining !== undefined ? (
          <CardContent>
            <p className="text-xs text-muted-foreground">
              {t("issueSearch.rateRemaining", {
                count: formatCompactNumber(
                  envelope.meta.rateLimitRemaining,
                  locale,
                ),
              })}
            </p>
          </CardContent>
        ) : null}
      </Card>

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-surface p-4">
        <label className="text-sm font-semibold" htmlFor="issue-result-sort">
          {t("issueSearch.sortLabel")}
        </label>
        <Select
          onValueChange={(value) => onSortChange?.(value as IssueSort)}
          value={sortBy}
        >
          <SelectTrigger id="issue-result-sort">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {searchFilterOptions.sorts.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {t(`issueSearch.sort.${option.value}`)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {relaxed ? (
        <Alert variant="warning">
          <AlertTitle>{t("search.relaxedTitle")}</AlertTitle>
        </Alert>
      ) : null}

      {hasPartialEvidence ? (
        <Alert variant="warning">
          <AlertTitle>{t("issueSearch.partialTitle")}</AlertTitle>
          <AlertDescription>
            {t("issueSearch.partialDescription")}{" "}
            {warnings.map((warning) => warning.message).join(" ")}
          </AlertDescription>
        </Alert>
      ) : (
        <Alert variant="success">
          <AlertTitle>{t("issueSearch.completeTitle")}</AlertTitle>
          <AlertDescription>
            <span className="inline-flex items-center gap-2">
              <Icon icon={SearchCheck} />
              {t("issueSearch.completeDescription")}
            </span>
          </AlertDescription>
        </Alert>
      )}

      <ol className="grid gap-5">
        {items.map((item, index) => {
          const rank = firstRank + index;
          return (
            <li key={`${item.repository.fullName}#${item.issue.number}`}>
              <RecommendationCard item={item} rank={rank} />
            </li>
          );
        })}
      </ol>

      <Card>
        <CardContent className="p-5 sm:p-6">
          <Pagination
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
