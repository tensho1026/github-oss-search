import { useMutation, useQueryClient } from "@tanstack/react-query";
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
import { createSavedSearch } from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

type Props =
  | { filters: IssueSearchRequest; searchType: "issue" }
  | { filters: RepositoryDiscoveryRequest; searchType: "repository" };

export function SaveSearchAction(props: Props) {
  const { t } = useI18n();
  const { markSessionExpired, session, signIn } = useAuth();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [formError, setFormError] = useState("");
  const [savedFilters, setSavedFilters] = useState("");
  const filterKey = JSON.stringify(props.filters);
  const mutation = useMutation({
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
      setName("");
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
    mutation.mutate({
      filters: props.filters,
      name: name.trim(),
      searchType: props.searchType,
    } as SavedSearchWriteRequest);
  }

  const authenticated = session?.authenticated && session.csrfToken;
  return (
    <>
      <Button onClick={() => setOpen(true)} size="small" variant="outline">
        {savedFilters === filterKey
          ? t("savedSearch.saved")
          : t("savedSearch.saveThis")}
      </Button>
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
              {mutation.error ? (
                <AccountRequestAlert error={mutation.error} />
              ) : null}
              <Button disabled={mutation.isPending} type="submit">
                {mutation.isPending
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
