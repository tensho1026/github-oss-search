import { Search, SlidersHorizontal } from "lucide-react";
import { useMemo, useState } from "react";
import { useLocation, useSearchParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { CopyLinkButton } from "../components/ui/copy-link-button";
import { FilterChipList } from "../components/ui/filter-chip-list";
import { Icon } from "../components/ui/icon";
import { usePreferredPageSize } from "../features/account/api/usePreferredPageSize";
import { SaveSearchAction } from "../features/account/components/SaveSearchAction";
import { useAuth } from "../features/auth/auth-context";
import { useIssueSearch } from "../features/issue-search/api/useIssueSearch";
import { IssueSearchForm } from "../features/issue-search/components/IssueSearchForm";
import { IssueSearchResults } from "../features/issue-search/components/IssueSearchResults";
import {
  IssueSearchBeforeState,
  IssueSearchErrorState,
  IssueSearchInvalidState,
  IssueSearchLoadingState,
} from "../features/issue-search/components/IssueSearchState";
import { emptyIssueSearchActions } from "../features/issue-search/model/search-empty-actions";
import { issueSearchFilterChips } from "../features/issue-search/model/search-filter-chips";
import {
  decodeSearchParams,
  encodeSearchParams,
  toIssueSearchRequest,
  validateSearchFilters,
  type IssueSort,
  type SearchFilters,
} from "../features/issue-search/model/search-filters";
import {
  appendSearchSelection,
  decodeSearchSelection,
  searchSelectionKey,
  toggleSearchSelection,
} from "../features/issue-search/model/search-selection";
import type { CompareReference } from "../features/issue-compare/model/compare-location";
import { useI18n } from "../shared/i18n/i18n-context";
import {
  appendSavedSearchId,
  decodeSavedSearchId,
} from "../shared/lib/saved-search-location";
import type { MessageKey } from "../shared/i18n/messages";

const emptyActionLabels: Record<
  ReturnType<typeof emptyIssueSearchActions>[number]["id"],
  MessageKey
> = {
  "clear-effort": "search.emptyClearEffort",
  "clear-frameworks": "search.emptyClearFrameworks",
  "clear-stars": "search.emptyClearStars",
  "raise-difficulty": "search.emptyRaiseDifficulty",
  "relax-recency": "search.emptyRelaxRecency",
};

export function IssueSearchPage() {
  const { session } = useAuth();
  const { t } = useI18n();
  const routeLocation = useLocation();
  const [searchParameters, setSearchParameters] = useSearchParams();
  const [mobileFormOpen, setMobileFormOpen] = useState(false);
  const preferredPageSize = usePreferredPageSize();
  const serializedSearch = searchParameters.toString();
  const selected = useMemo(
    () => decodeSearchSelection(new URLSearchParams(serializedSearch)),
    [serializedSearch],
  );
  const savedSearchId = useMemo(
    () => decodeSavedSearchId(new URLSearchParams(serializedSearch)),
    [serializedSearch],
  );
  const location = useMemo(() => {
    const decoded = decodeSearchParams(new URLSearchParams(serializedSearch), {
      perPage: preferredPageSize,
    });
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
  }, [preferredPageSize, serializedSearch, session]);
  const query = useIssueSearch(location);

  function writeLocation(
    filters: SearchFilters,
    options: {
      savedSearchId?: string;
      selected?: readonly CompareReference[];
      shouldSearch?: boolean;
    } = {},
  ) {
    const parameters = encodeSearchParams(
      filters,
      options.shouldSearch ?? true,
    );
    appendSearchSelection(parameters, options.selected ?? selected);
    appendSavedSearchId(
      parameters,
      options.savedSearchId === undefined
        ? savedSearchId
        : options.savedSearchId,
    );
    setSearchParameters(parameters);
  }

  function submit(filters: SearchFilters) {
    writeLocation(filters);
    setMobileFormOpen(false);
  }

  function changePage(page: number) {
    writeLocation({ ...location.filters, page });
  }

  function changeSort(sortBy: IssueSort) {
    writeLocation({ ...location.filters, page: 1, sortBy });
  }

  const chips = issueSearchFilterChips(location.filters);
  const emptyActions = location.valid
    ? emptyIssueSearchActions(location.filters).map((action) => ({
        href: `/search?${encodeSearchParams(action.filters).toString()}`,
        label: t(emptyActionLabels[action.id]),
      }))
    : [];

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
        emptyActions={emptyActions}
        envelope={query.data[0]}
        isFetching={query.isFetching}
        onPageChange={changePage}
        onSortChange={changeSort}
        sortBy={location.filters.sortBy}
        selectedItems={selected}
        onClearSelection={() =>
          writeLocation(location.filters, { selected: [] })
        }
        onSelectionChange={(item, nextSelected) => {
          const reference = {
            issueNumber: item.issue.number,
            owner: item.repository.owner,
            repository: item.repository.name,
          };
          writeLocation(location.filters, {
            selected: nextSelected
              ? toggleSearchSelection(selected, reference)
              : selected.filter(
                  (candidate) =>
                    searchSelectionKey(candidate) !==
                    searchSelectionKey(reference),
                ),
          });
        }}
        relaxed={query.data[1]}
        returnTo={`${routeLocation.pathname}${routeLocation.search}`}
        skills={[...location.filters.languages, ...location.filters.frameworks]}
      />
    );
  } else {
    resultContent = <IssueSearchLoadingState />;
  }

  const showForm = !location.shouldSearch || mobileFormOpen;

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
          <div className="mt-5 flex flex-wrap gap-2">
            <SaveSearchAction
              filters={toIssueSearchRequest(location.filters)}
              savedSearchId={savedSearchId}
              searchType="issue"
            />
            <CopyLinkButton />
          </div>
        ) : null}
      </header>

      {chips.length > 0 ? (
        <div className="mt-6">
          <FilterChipList
            chips={chips.map((chip) => ({
              id: chip.id,
              label: chip.label,
              onRemove: () =>
                writeLocation(chip.nextFilters, {
                  shouldSearch: location.shouldSearch && location.valid,
                }),
            }))}
          />
        </div>
      ) : null}

      {location.shouldSearch ? (
        <div className="mt-6 lg:hidden">
          <Button
            aria-expanded={showForm}
            onClick={() => setMobileFormOpen((open) => !open)}
            size="small"
            variant="outline"
          >
            {showForm ? t("search.hideFilters") : t("search.showFilters")}
          </Button>
        </div>
      ) : null}

      <div className="mt-9 grid items-start gap-6 lg:grid-cols-[minmax(20rem,0.82fr)_minmax(0,1.18fr)]">
        <Card
          className={`overflow-hidden lg:sticky lg:top-24 ${showForm ? "" : "max-lg:hidden"}`}
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
              sessionUsername={session?.user?.login}
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
