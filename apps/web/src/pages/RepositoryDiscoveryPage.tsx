import { BookOpenCheck, SlidersHorizontal } from "lucide-react";
import { useMemo, useState } from "react";
import { useSearchParams } from "react-router";

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
import { useRepositoryDiscovery } from "../features/repository-discovery/api/useRepositoryDiscovery";
import { RepositoryDiscoveryForm } from "../features/repository-discovery/components/RepositoryDiscoveryForm";
import { RepositoryDiscoveryResults } from "../features/repository-discovery/components/RepositoryDiscoveryResults";
import {
  RepositoryDiscoveryBeforeState,
  RepositoryDiscoveryErrorState,
  RepositoryDiscoveryInvalidState,
  RepositoryDiscoveryLoadingState,
} from "../features/repository-discovery/components/RepositoryDiscoveryState";
import {
  emptyRepositorySearchActions,
  repositoryFilterChips,
} from "../features/repository-discovery/model/repository-filter-chips";
import {
  decodeRepositorySearchParams,
  encodeRepositorySearchParams,
  toRepositoryDiscoveryRequest,
  type RepositoryFilters,
} from "../features/repository-discovery/model/repository-filters";
import { useI18n } from "../shared/i18n/i18n-context";
import type { MessageKey } from "../shared/i18n/messages";
import {
  appendSavedSearchId,
  decodeSavedSearchId,
} from "../shared/lib/saved-search-location";

const emptyActionLabels: Record<
  ReturnType<typeof emptyRepositorySearchActions>[number]["id"],
  MessageKey
> = {
  "clear-stars": "search.emptyClearStars",
  "clear-technologies": "search.emptyClearTechnologies",
  "raise-difficulty": "search.emptyRaiseDifficulty",
  "relax-recency": "search.emptyRelaxRecency",
};

export function RepositoryDiscoveryPage() {
  const { t } = useI18n();
  const [searchParameters, setSearchParameters] = useSearchParams();
  const [mobileFormOpen, setMobileFormOpen] = useState(false);
  const preferredPageSize = usePreferredPageSize();
  const serializedSearch = searchParameters.toString();
  const savedSearchId = useMemo(
    () => decodeSavedSearchId(new URLSearchParams(serializedSearch)),
    [serializedSearch],
  );
  const location = useMemo(
    () =>
      decodeRepositorySearchParams(new URLSearchParams(serializedSearch), {
        perPage: preferredPageSize,
      }),
    [preferredPageSize, serializedSearch],
  );
  const query = useRepositoryDiscovery(location);

  function writeLocation(
    filters: RepositoryFilters,
    options: { savedSearchId?: string; shouldSearch?: boolean } = {},
  ) {
    const parameters = encodeRepositorySearchParams(
      filters,
      options.shouldSearch ?? true,
    );
    appendSavedSearchId(
      parameters,
      options.savedSearchId === undefined
        ? savedSearchId
        : options.savedSearchId,
    );
    setSearchParameters(parameters);
  }

  function submit(filters: RepositoryFilters) {
    writeLocation(filters);
    setMobileFormOpen(false);
  }

  function changePage(page: number) {
    writeLocation({ ...location.filters, page });
  }

  const chips = repositoryFilterChips(location.filters);
  const emptyActions = emptyRepositorySearchActions(location.filters).map(
    (action) => ({
      href: `/repositories?${encodeRepositorySearchParams(action.filters).toString()}`,
      label: t(emptyActionLabels[action.id]),
    }),
  );

  let resultContent;
  if (location.shouldSearch && !location.valid) {
    resultContent = <RepositoryDiscoveryInvalidState />;
  } else if (!location.shouldSearch) {
    resultContent = <RepositoryDiscoveryBeforeState />;
  } else if (query.isPending) {
    resultContent = <RepositoryDiscoveryLoadingState />;
  } else if (query.error) {
    resultContent = (
      <RepositoryDiscoveryErrorState
        error={query.error}
        isFetching={query.isFetching}
        onRetry={() => {
          void query.refetch();
        }}
      />
    );
  } else if (query.data) {
    resultContent = (
      <RepositoryDiscoveryResults
        contributorTechnologies={[
          ...location.filters.languages,
          ...location.filters.technologies,
        ]}
        emptyActions={emptyActions}
        envelope={query.data[0]}
        filters={location.filters}
        isFetching={query.isFetching}
        onPageChange={changePage}
        relaxed={query.data[1]}
      />
    );
  } else {
    resultContent = <RepositoryDiscoveryLoadingState />;
  }

  const showForm = !location.shouldSearch || mobileFormOpen;

  return (
    <div className="mx-auto w-full max-w-7xl px-5 py-10 sm:px-8 sm:py-14 lg:px-10">
      <header className="max-w-3xl">
        <Badge variant="accent">
          <Icon icon={BookOpenCheck} />
          {t("repository.badge")}
        </Badge>
        <h1 className="mt-5 text-4xl font-semibold tracking-[-0.055em] text-balance sm:text-5xl">
          {t("repository.title")}
        </h1>
        <p className="mt-5 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg">
          {t("repository.description")}
        </p>
        {location.shouldSearch && location.valid ? (
          <div className="mt-5 flex flex-wrap gap-2">
            <SaveSearchAction
              filters={toRepositoryDiscoveryRequest(location.filters)}
              savedSearchId={savedSearchId}
              searchType="repository"
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
        <div className="mt-6 xl:hidden">
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

      <div className="mt-9 grid items-start gap-6 xl:grid-cols-[minmax(24rem,0.86fr)_minmax(0,1.14fr)]">
        <Card
          className={`overflow-hidden xl:sticky xl:top-24 ${showForm ? "" : "max-xl:hidden"}`}
          id="repository-filters"
        >
          <CardHeader className="border-b border-border bg-muted/35">
            <span className="grid size-11 place-items-center rounded-xl bg-accent-soft text-accent-soft-foreground">
              <Icon className="size-5" icon={SlidersHorizontal} />
            </span>
            <CardTitle className="mt-2">{t("repository.criteria")}</CardTitle>
            <CardDescription>
              {t("repository.criteriaDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent className="p-5 sm:p-6">
            <RepositoryDiscoveryForm
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
