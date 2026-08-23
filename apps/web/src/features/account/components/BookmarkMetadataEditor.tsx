import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { ApiError } from "../../../shared/api/client";
import type { Bookmark } from "../../../shared/api/generated";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { queryKeys } from "../../../shared/query/query-keys";
import { updateBookmarkMetadata } from "../api/account";
import { AccountRequestAlert } from "./AccountRequestAlert";

export function BookmarkMetadataEditor({
  bookmark,
  csrfToken,
  onSessionExpired,
}: {
  bookmark: Bookmark;
  csrfToken: string;
  onSessionExpired: () => Promise<void>;
}) {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState(false);
  const [note, setNote] = useState(bookmark.note);
  const [collection, setCollection] = useState(bookmark.collection);
  const [tags, setTags] = useState(bookmark.tags.join(", "));
  const mutation = useMutation({
    mutationFn: () =>
      updateBookmarkMetadata(
        bookmark.id,
        {
          collection,
          note,
          tags: tags
            .split(",")
            .map((tag) => tag.trim())
            .filter(Boolean),
          version: bookmark.version,
        },
        csrfToken,
      ),
    async onError(error) {
      if (error instanceof ApiError && error.status === 401) {
        await onSessionExpired();
      }
    },
    async onSuccess() {
      setEditing(false);
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.bookmarks,
      });
    },
  });

  if (!editing) {
    return (
      <div className="basis-full border-t border-border pt-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="grid gap-2 text-sm">
            {bookmark.collection ? (
              <p>
                <strong>{t("bookmarkMeta.collection")}:</strong>{" "}
                {bookmark.collection}
              </p>
            ) : null}
            {bookmark.tags.length > 0 ? (
              <div className="flex flex-wrap gap-2">
                {bookmark.tags.map((tag) => (
                  <Badge key={tag} variant="accent">
                    {tag}
                  </Badge>
                ))}
              </div>
            ) : null}
            {bookmark.note ? (
              <p className="whitespace-pre-wrap text-muted-foreground">
                {bookmark.note}
              </p>
            ) : (
              <p className="text-muted-foreground">{t("bookmarkMeta.empty")}</p>
            )}
          </div>
          <Button onClick={() => setEditing(true)} size="small" variant="ghost">
            {t("bookmarkMeta.edit")}
          </Button>
        </div>
      </div>
    );
  }
  return (
    <div className="grid basis-full gap-3 border-t border-border pt-4">
      <label className="grid gap-1 text-sm font-medium">
        {t("bookmarkMeta.collection")}
        <Input
          maxLength={80}
          onChange={(event) => setCollection(event.target.value)}
          value={collection}
        />
      </label>
      <label className="grid gap-1 text-sm font-medium">
        {t("bookmarkMeta.tags")}
        <Input
          maxLength={339}
          onChange={(event) => setTags(event.target.value)}
          placeholder={t("bookmarkMeta.tagsHint")}
          value={tags}
        />
      </label>
      <label className="grid gap-1 text-sm font-medium">
        {t("bookmarkMeta.note")}
        <textarea
          className="min-h-28 rounded-xl border border-input bg-surface px-3 py-2"
          maxLength={500}
          onChange={(event) => setNote(event.target.value)}
          value={note}
        />
      </label>
      <div className="flex gap-2">
        <Button
          disabled={mutation.isPending}
          onClick={() => mutation.mutate()}
          size="small"
        >
          {mutation.isPending
            ? t("bookmarkMeta.saving")
            : t("bookmarkMeta.save")}
        </Button>
        <Button onClick={() => setEditing(false)} size="small" variant="ghost">
          {t("bookmarkMeta.cancel")}
        </Button>
      </div>
      {mutation.error ? <AccountRequestAlert error={mutation.error} /> : null}
    </div>
  );
}
