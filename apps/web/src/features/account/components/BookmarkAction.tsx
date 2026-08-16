import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { Button } from "../../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "../../../components/ui/dialog";
import { ApiError } from "../../../shared/api/client";
import type { BookmarkWriteRequest } from "../../../shared/api/generated";
import { queryKeys } from "../../../shared/query/query-keys";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { useAuth } from "../../auth/auth-context";
import { upsertBookmark } from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

export function BookmarkAction({ request }: { request: BookmarkWriteRequest }) {
  const { t } = useI18n();
  const { markSessionExpired, session, signIn } = useAuth();
  const queryClient = useQueryClient();
  const [promptOpen, setPromptOpen] = useState(false);
  const [savedTarget, setSavedTarget] = useState("");
  const targetKey = JSON.stringify(request);
  const mutation = useMutation({
    mutationFn: () => {
      if (!session?.authenticated || !session.csrfToken) {
        throw new ApiError({
          code: "AUTHENTICATION_REQUIRED",
          message: "Authentication is required.",
          status: 401,
        });
      }
      return upsertBookmark(request, session.csrfToken);
    },
    async onError(error) {
      if (error instanceof ApiError && error.status === 401) {
        await markSessionExpired();
        setPromptOpen(true);
      }
    },
    async onSuccess() {
      setSavedTarget(targetKey);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.bookmarks,
      });
    },
  });

  const authenticated = session?.authenticated && session.csrfToken;
  return (
    <>
      <Button
        disabled={mutation.isPending}
        onClick={() => {
          if (authenticated) {
            mutation.mutate();
          } else {
            setPromptOpen(true);
          }
        }}
        variant="outline"
      >
        {mutation.isPending
          ? t("bookmark.saving")
          : savedTarget === targetKey
            ? t("bookmark.saved")
            : t("bookmark.save")}
      </Button>
      <Dialog onOpenChange={setPromptOpen} open={promptOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("bookmark.dialogTitle")}</DialogTitle>
            <DialogDescription>
              {t("bookmark.dialogDescription")}
            </DialogDescription>
          </DialogHeader>
          {session?.configured === false ? (
            <p className="text-sm text-muted-foreground">
              {t("workspace.notConfigured")}
            </p>
          ) : (
            <Button
              onClick={() => {
                signIn();
              }}
            >
              {t("workspace.signInGitHub")}
            </Button>
          )}
        </DialogContent>
      </Dialog>
      {mutation.error &&
      (!(mutation.error instanceof ApiError) ||
        mutation.error.status !== 401) ? (
        <AccountRequestAlert error={mutation.error} />
      ) : null}
    </>
  );
}
