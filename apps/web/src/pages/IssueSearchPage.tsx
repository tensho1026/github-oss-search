import { Search, SlidersHorizontal } from "lucide-react";
import { useMemo } from "react";
import { useSearchParams } from "react-router";

import { Badge } from "../components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Icon } from "../components/ui/icon";
import { useAuth } from "../features/auth/auth-context";
import { useIssueSearch } from "../features/issue-search/api/useIssueSearch";
import { SaveSearchAction } from "../features/account/components/SaveSearchAction";
import { IssueSearchForm } from "../features/issue-search/components/IssueSearchForm";
import { IssueSearchResults } from "../features/issue-search/components/IssueSearchResults";
import {
  IssueSearchBeforeState,
  IssueSearchErrorState,
  IssueSearchInvalidState,
  IssueSearchLoadingState,
} from "../features/issue-search/components/IssueSearchState";
import {
  decodeSearchParams,
  encodeSearchParams,
  toIssueSearchRequest,
  validateSearchFilters,
  type SearchFilters,
  type IssueSort,
} from "../features/issue-search/model/search-filters";
import { useI18n } from "../shared/i18n/i18n-context";

export function IssueSearchPage() {
  const { session } = useAuth();
  const { t } = useI18n();
  const [searchParameters, setSearchParameters] = useSearchParams();
  const serializedSearch = searchParameters.toString();
  const location = useMemo(() => {
    const decoded = decodeSearchParams(new URLSearchParams(serializedSearch));
    if (
      decoded.filters.username ||
      !session?.authenticated ||
      !session.user?.login
    ) {
      return decoded;
    }
    const filters = { ...decoded.filters, username: session.user.login };
    const errors = validateSearchFilters(filters);
    return {
      ...decoded,
      errors,
      filters,
      valid: Object.keys(errors).length === 0,
    };
  }, [serializedSearch, session]);
  const query = useIssueSearch(location);

  function submit(filters: SearchFilters) {
    setSearchParameters(encodeSearchParams(filters));
  }

  function changePage(page: number) {
    setSearchParameters(
      encodeSearchParams({
        ...location.filters,
        page,
      }),
    );
  }

  function changeSort(sortBy: IssueSort) {
    setSearchParameters(
      encodeSearchParams({ ...location.filters, page: 1, sortBy }),
    );
  }

  let resultContent;
  if (location.shouldSearch && !location.valid) {
    resultContent = <IssueSearchInvalidState />;
  } else if (!location.shouldSearch) {
    resultContent = <IssueSearchBeforeState />;
  } else if (query.isPending) {
    resultContent = <IssueSearchLoadingState />;
  } else if (query.error) {
    resultContent = (
      <IssueSearchErrorState
        error={query.error}
        isFetching={query.isFetching}
        onRetry={() => {
          void query.refetch();
        }}
      />
    );
  } else if (query.data) {
    resultContent = (
      <IssueSearchResults
        envelope={query.data[0]}
        isFetching={query.isFetching}
        relaxed={query.data[1]}
        onPageChange={changePage}
        onSortChange={changeSort}
        sortBy={location.filters.sortBy}
      />
    );
  } else {
    resultContent = <IssueSearchLoadingState />;
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-5 py-10 sm:px-8 sm:py-14 lg:px-10">
      <header className="max-w-3xl">
        <Badge variant="accent">
          <Icon icon={Search} />
          {t(
            session?.authenticated
              ? "issueSearch.personalizedBadge"
              : "issueSearch.badge",
          )}
        </Badge>
        <h1 className="mt-5 text-4xl font-semibold tracking-[-0.055em] text-balance sm:text-5xl">
          {t("issueSearch.title")}
        </h1>
        <p className="mt-5 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
          {t("issueSearch.description")}
        </p>
        {location.shouldSearch && location.valid ? (
          <div className="mt-5">
            <SaveSearchAction
              filters={toIssueSearchRequest(location.filters)}
              searchType="issue"
            />
          </div>
        ) : null}
      </header>

      <div className="mt-9 grid items-start gap-6 lg:grid-cols-[minmax(20rem,0.82fr)_minmax(0,1.18fr)]">
        <Card
          className="overflow-hidden lg:sticky lg:top-24"
          id="search-filters"
        >
          <CardHeader className="border-b border-border bg-muted/35">
            <span className="grid size-11 place-items-center rounded-xl bg-accent-soft text-accent-soft-foreground">
              <Icon className="size-5" icon={SlidersHorizontal} />
            </span>
            <CardTitle className="mt-2">{t("issueSearch.criteria")}</CardTitle>
            <CardDescription>
              {t("issueSearch.criteriaDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent className="p-5 sm:p-6">
            <IssueSearchForm
              defaultValues={location.filters}
              disabled={query.isFetching}
              locationErrors={
                location.shouldSearch ? location.errors : undefined
              }
              onSubmit={submit}
            />
          </CardContent>
        </Card>

        <div aria-live="polite" aria-relevant="additions text">
          {resultContent}
        </div>
      </div>
    </div>
  );
}
