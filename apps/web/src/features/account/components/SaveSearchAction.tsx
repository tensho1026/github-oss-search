import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";

import { Button } from "../../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "../../../components/ui/dialog";
import { Field } from "../../../components/ui/field";
import { Input } from "../../../components/ui/input";
import { ApiError } from "../../../shared/api/client";
import type {
  IssueSearchRequest,
  RepositoryDiscoveryRequest,
  SavedSearchWriteRequest,
} from "../../../shared/api/generated";
import { queryKeys } from "../../../shared/query/query-keys";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { useAuth } from "../../auth/auth-context";
import {
  createDefaultSearchFilters,
  type SearchFilters,
} from "../../issue-search/model/search-filters";
import {
  suggestedIssueSearchName,
  suggestedRepositorySearchName,
} from "../../issue-search/model/saved-search-name";
import { createDefaultRepositoryFilters } from "../../repository-discovery/model/repository-filters";
import {
  createSavedSearch,
  listSavedSearches,
  updateSavedSearch,
} from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

type Props = (
  | { filters: IssueSearchRequest; searchType: "issue" }
  | { filters: RepositoryDiscoveryRequest; searchType: "repository" }
) & { savedSearchId?: string };

export function SaveSearchAction(props: Props) {
  const { t } = useI18n();
  const { markSessionExpired, session, signIn } = useAuth();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState(suggestedName(props));
  const [formError, setFormError] = useState("");
  const [savedFilters, setSavedFilters] = useState("");
  const filterKey = JSON.stringify(props.filters);
  const savedSearchQuery = useQuery({
    enabled: Boolean(
      session?.authenticated && session.csrfToken && props.savedSearchId,
    ),
    queryFn: ({ signal }) => listSavedSearches(1, 50, signal),
    queryKey: queryKeys.account.savedSearchPage(1, 50),
  });
  const currentSavedSearch = savedSearchQuery.data?.data.items.find(
    (search) => search.id === props.savedSearchId,
  );
  const createMutation = useMutation({
    mutationFn: (request: SavedSearchWriteRequest) => {
      if (!session?.authenticated || !session.csrfToken) {
        throw new ApiError({
          code: "AUTHENTICATION_REQUIRED",
          message: "Authentication is required.",
          status: 401,
        });
      }
      return createSavedSearch(request, session.csrfToken);
    },
    async onError(error) {
      if (error instanceof ApiError && error.status === 401) {
        await markSessionExpired();
      }
    },
    async onSuccess() {
      setOpen(false);
      setSavedFilters(filterKey);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.savedSearches,
      });
    },
  });
  const updateMutation = useMutation({
    mutationFn: () => {
      if (
        !session?.authenticated ||
        !session.csrfToken ||
        !currentSavedSearch
      ) {
        throw new ApiError({
          code: "AUTHENTICATION_REQUIRED",
          message: "Authentication is required.",
          status: 401,
        });
      }
      return updateSavedSearch(
        currentSavedSearch.id,
        currentSavedSearch.searchType === "issue"
          ? {
              filters: props.filters as IssueSearchRequest,
              name: currentSavedSearch.name,
              searchType: "issue",
              version: currentSavedSearch.version,
            }
          : {
              filters: props.filters,
              name: currentSavedSearch.name,
              searchType: "repository",
              version: currentSavedSearch.version,
            },
        session.csrfToken,
      );
    },
    async onError(error) {
      if (error instanceof ApiError && error.status === 401) {
        await markSessionExpired();
      }
    },
    async onSuccess() {
      setSavedFilters(filterKey);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.savedSearches,
      });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!name.trim()) {
      setFormError(t("savedSearch.nameRequired"));
      return;
    }
    setFormError("");
    createMutation.mutate({
      filters: props.filters,
      name: name.trim(),
      searchType: props.searchType,
    } as SavedSearchWriteRequest);
  }

  const authenticated = session?.authenticated && session.csrfToken;
  return (
    <>
      <Button
        onClick={() => {
          setName((current) => current.trim() || suggestedName(props));
          setOpen(true);
        }}
        size="small"
        variant="outline"
      >
        {savedFilters === filterKey
          ? t("savedSearch.saved")
          : t("savedSearch.saveThis")}
      </Button>
      {currentSavedSearch ? (
        <Button
          disabled={updateMutation.isPending || savedFilters === filterKey}
          onClick={() => updateMutation.mutate()}
          size="small"
          variant="secondary"
        >
          {updateMutation.isPending
            ? t("savedSearch.updating")
            : updateMutation.isSuccess && savedFilters === filterKey
              ? t("savedSearch.updated")
              : t("savedSearch.updateCurrent")}
        </Button>
      ) : null}
      <Dialog onOpenChange={setOpen} open={open}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {authenticated
                ? t("savedSearch.nameTitle")
                : t("savedSearch.filtersTitle")}
            </DialogTitle>
            <DialogDescription>
              {authenticated
                ? t("savedSearch.authDescription")
                : t("savedSearch.guestDescription")}
            </DialogDescription>
          </DialogHeader>
          {authenticated ? (
            <form className="grid gap-4" onSubmit={submit}>
              <Field
                error={formError || undefined}
                htmlFor={`save-${props.searchType}-search-name`}
                label={t("savedSearch.name")}
              >
                <Input
                  id={`save-${props.searchType}-search-name`}
                  maxLength={80}
                  onChange={(event) => setName(event.target.value)}
                  value={name}
                />
              </Field>
              {createMutation.error ? (
                <AccountRequestAlert error={createMutation.error} />
              ) : null}
              {updateMutation.error ? (
                <AccountRequestAlert error={updateMutation.error} />
              ) : null}
              <Button disabled={createMutation.isPending} type="submit">
                {createMutation.isPending
                  ? t("savedSearch.saving")
                  : t("savedSearch.save")}
              </Button>
            </form>
          ) : session?.configured === false ? (
            <p className="text-sm text-muted-foreground">
              {t("workspace.notConfigured")}
            </p>
          ) : (
            <Button onClick={() => signIn()}>
              {t("workspace.signInGitHub")}
            </Button>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}

function suggestedName(props: Props): string {
  if (props.searchType === "issue") {
    const filters = props.filters;
    const next: SearchFilters = {
      ...createDefaultSearchFilters(filters.username),
      excludeArchived: filters.excludeArchived ?? true,
      frameworks: filters.frameworks ?? [],
      includeDocumentation: filters.includeDocumentation ?? false,
      includeEnglish: filters.includeEnglish ?? true,
      labels: filters.labels ?? ["good first issue", "help wanted"],
      languages: filters.languages ?? [],
      maximumDifficulty: filters.maximumDifficulty ?? 3,
      maximumEffort: filters.maximumEffort ?? "",
      minimumStars: filters.minimumStars ?? 10,
      updatedWithinDays: filters.updatedWithinDays ?? 180,
    };
    return suggestedIssueSearchName(next);
  }
  const filters = props.filters;
  return suggestedRepositorySearchName({
    ...createDefaultRepositoryFilters(),
    categories: filters.categories ?? [],
    languages: filters.languages ?? [],
    licenses: filters.licenses ?? [],
    minimumStars: filters.minimumStars ?? 10,
    technologies: filters.technologies ?? [],
  });
}
