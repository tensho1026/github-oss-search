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
import type { IssueClaimWriteRequest } from "../../../shared/api/generated";
import { queryKeys } from "../../../shared/query/query-keys";
import { useAuth } from "../../auth/auth-context";
import { upsertIssueClaim } from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

export function IssueClaimAction({
  request,
}: {
  request: IssueClaimWriteRequest;
}) {
  const { markSessionExpired, session, signIn } = useAuth();
  const queryClient = useQueryClient();
  const [promptOpen, setPromptOpen] = useState(false);
  const [saved, setSaved] = useState(false);
  const mutation = useMutation({
    mutationFn: () => {
      if (!session?.authenticated || !session.csrfToken) {
        throw new ApiError({
          code: "AUTHENTICATION_REQUIRED",
          message: "Authentication is required.",
          status: 401,
        });
      }
      return upsertIssueClaim(request, session.csrfToken);
    },
    async onError(error) {
      if (error instanceof ApiError && error.status === 401) {
        await markSessionExpired();
        setPromptOpen(true);
      }
    },
    async onSuccess() {
      setSaved(true);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.issueClaims,
      });
    },
  });
  const authenticated = session?.authenticated && session.csrfToken;

  return (
    <>
      <Button
        disabled={mutation.isPending}
        onClick={() => {
          if (authenticated) mutation.mutate();
          else setPromptOpen(true);
        }}
        variant="outline"
      >
        {mutation.isPending
          ? "Adding…"
          : saved
            ? "Added to task board"
            : "Try this issue"}
      </Button>
      <Dialog onOpenChange={setPromptOpen} open={promptOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add this issue to your task board</DialogTitle>
            <DialogDescription>
              This is your private IssueScout progress state. It does not assign
              the GitHub issue, post a comment, or imply maintainer approval.
            </DialogDescription>
          </DialogHeader>
          {session?.configured === false ? (
            <p className="text-sm text-muted-foreground">
              Sign-in is not configured in this environment.
            </p>
          ) : (
            <Button onClick={() => signIn()}>Sign in with GitHub</Button>
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
