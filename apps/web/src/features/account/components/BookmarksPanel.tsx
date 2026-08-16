import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { ApiError } from "../../../shared/api/client";
import { appRoutes, externalLinks } from "../../../shared/config/app-config";
import { queryKeys } from "../../../shared/query/query-keys";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { deleteBookmark, listBookmarks } from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

type Props = {
  csrfToken: string;
  onSessionExpired: () => Promise<void>;
};

export function BookmarksPanel({ csrfToken, onSessionExpired }: Props) {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const query = useQuery({
    queryFn: ({ signal }) => listBookmarks(signal),
    queryKey: queryKeys.account.bookmarks,
  });

  async function handleError(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      await onSessionExpired();
    }
  }

  const remove = useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) =>
      deleteBookmark(id, version, csrfToken),
    onError: handleError,
    async onSuccess() {
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.bookmarks,
      });
    },
  });

  return (
    <section aria-labelledby="bookmarks-heading" className="grid gap-5">
      <h2 className="sr-only" id="bookmarks-heading">
        {t("workspace.bookmarks")}
      </h2>
      {remove.error ? <AccountRequestAlert error={remove.error} /> : null}
      {query.error ? <AccountRequestAlert error={query.error} /> : null}

      {query.isPending ? (
        <Card>
          <CardContent
            className="p-6 text-sm text-muted-foreground"
            role="status"
          >
            {t("bookmarks.loading")}
          </CardContent>
        </Card>
      ) : query.data?.data.items.length === 0 ? (
        <Card>
          <CardContent className="grid justify-items-center gap-3 p-8 text-center">
            <p className="font-semibold">{t("bookmarks.empty")}</p>
            <p className="max-w-lg text-sm text-muted-foreground">
              {t("bookmarks.emptyDescription")}
            </p>
          </CardContent>
        </Card>
      ) : (
        <ul className="grid gap-3">
          {query.data?.data.items.map((bookmark) => {
            const label = `${bookmark.repositoryOwner}/${bookmark.repositoryName}${
              bookmark.issueNumber ? `#${bookmark.issueNumber}` : ""
            }`;
            const publicPath = bookmark.issueNumber
              ? appRoutes.issue(
                  bookmark.repositoryOwner,
                  bookmark.repositoryName,
                  bookmark.issueNumber,
                )
              : undefined;
            const externalPath = bookmark.issueNumber
              ? externalLinks.gitHubIssue(
                  bookmark.repositoryOwner,
                  bookmark.repositoryName,
                  bookmark.issueNumber,
                )
              : externalLinks.gitHubRepository(
                  bookmark.repositoryOwner,
                  bookmark.repositoryName,
                );
            return (
              <li key={bookmark.id}>
                <Card>
                  <CardContent className="flex flex-wrap items-center justify-between gap-4 p-5">
                    <div className="min-w-0">
                      <p className="truncate font-mono text-sm font-semibold">
                        {label}
                      </p>
                      <div className="mt-2 flex flex-wrap gap-2">
                        <Badge variant="neutral">{bookmark.targetType}</Badge>
                        <Badge variant="warning">
                          {t("bookmarks.upstream", {
                            state: bookmark.upstreamState,
                          })}
                        </Badge>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {publicPath ? (
                        <Button asChild size="small" variant="outline">
                          <Link to={publicPath}>
                            {t("bookmarks.revalidate")}
                          </Link>
                        </Button>
                      ) : (
                        <Button asChild size="small" variant="outline">
                          <a
                            href={externalPath}
                            rel="noreferrer"
                            target="_blank"
                          >
                            {t("bookmarks.openGitHub")}
                          </a>
                        </Button>
                      )}
                      <Button
                        aria-label={t("bookmarks.deleteLabel", { label })}
                        disabled={remove.isPending}
                        onClick={() =>
                          remove.mutate({
                            id: bookmark.id,
                            version: bookmark.version,
                          })
                        }
                        size="small"
                        variant="danger"
                      >
                        {t("bookmarks.delete")}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
