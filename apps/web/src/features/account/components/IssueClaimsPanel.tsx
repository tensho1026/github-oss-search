import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { Link } from "react-router";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { Input } from "../../../components/ui/input";
import { ApiError } from "../../../shared/api/client";
import type {
  IssueClaim,
  IssueClaimStatus,
  IssueClaimUpdateRequest,
} from "../../../shared/api/generated";
import { appRoutes, externalLinks } from "../../../shared/config/app-config";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { queryKeys } from "../../../shared/query/query-keys";
import {
  deleteIssueClaim,
  listIssueClaims,
  updateIssueClaim,
} from "../api/account";
import { claimMoveRequest } from "../model/claim-board";
import { AccountRequestAlert } from "./AccountRequestAlert";
import { ReferenceObservationButton } from "./ReferenceObservationButton";

type Props = {
  csrfToken: string;
  onSessionExpired: () => Promise<void>;
};

const statusOptions: IssueClaimStatus[] = [
  "not_started",
  "researching",
  "implementing",
  "pr_submitted",
  "merged",
];

export function IssueClaimsPanel({ csrfToken, onSessionExpired }: Props) {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [filter, setFilter] = useState<"active" | "archived" | "all">("active");
  const [view, setView] = useState<"kanban" | "list">("kanban");
  const [draggedClaimID, setDraggedClaimID] = useState<string | null>(null);
  const [moveError, setMoveError] = useState("");
  const query = useQuery({
    queryFn: ({ signal }) => listIssueClaims(signal),
    queryKey: queryKeys.account.issueClaims,
  });

  async function handleError(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      await onSessionExpired();
    }
  }
  async function refresh() {
    await queryClient.invalidateQueries({
      queryKey: queryKeys.account.issueClaims,
    });
  }
  const update = useMutation({
    mutationFn: ({
      id,
      request,
    }: {
      id: string;
      request: IssueClaimUpdateRequest;
    }) => updateIssueClaim(id, request, csrfToken),
    onError: handleError,
    onSuccess: refresh,
  });
  const remove = useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) =>
      deleteIssueClaim(id, version, csrfToken),
    onError: handleError,
    onSuccess: refresh,
  });
  const claims = useMemo(() => {
    const items = query.data?.data.items ?? [];
    if (filter === "all") return items;
    return items.filter((claim) => claim.archived === (filter === "archived"));
  }, [filter, query.data]);
  const summary = query.data?.data.summary;
  const statusLabels: Record<IssueClaimStatus, string> = {
    implementing: t("claims.implementing"),
    merged: t("claims.merged"),
    not_started: t("claims.notStarted"),
    pr_submitted: t("claims.prSubmitted"),
    researching: t("claims.researching"),
  };
  const moveClaim = (claim: IssueClaim, status: IssueClaimStatus) => {
    setMoveError("");
    const move = claimMoveRequest(claim, status);
    if (!move.ok && move.reason === "pull_request_required") {
      setMoveError(t("claims.movePrRequired"));
      return;
    }
    if (!move.ok) return;
    update.mutate({
      id: claim.id,
      request: move.request,
    });
  };

  return (
    <section aria-labelledby="issue-claims-heading" className="grid gap-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold" id="issue-claims-heading">
            {t("claims.title")}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("claims.description")}
          </p>
        </div>
        <div className="flex flex-wrap items-end gap-3">
          <div
            aria-label={t("claims.view")}
            className="flex rounded-xl border border-border p-1"
            role="group"
          >
            {(["kanban", "list"] as const).map((option) => (
              <Button
                aria-pressed={view === option}
                key={option}
                onClick={() => setView(option)}
                size="small"
                variant={view === option ? "primary" : "ghost"}
              >
                {t(option === "kanban" ? "claims.kanban" : "claims.list")}
              </Button>
            ))}
          </div>
          <label className="grid gap-1 text-sm font-medium">
            {t("claims.show")}
            <select
              className="min-h-11 rounded-xl border border-input bg-surface px-3"
              onChange={(event) =>
                setFilter(event.target.value as "active" | "archived" | "all")
              }
              value={filter}
            >
              <option value="active">{t("claims.active")}</option>
              <option value="archived">{t("claims.archived")}</option>
              <option value="all">{t("claims.all")}</option>
            </select>
          </label>
        </div>
      </div>

      {summary ? (
        <dl className="grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-7">
          {[
            [t("claims.total"), summary.total],
            [t("claims.notStarted"), summary.notStarted],
            [t("claims.researching"), summary.researching],
            [t("claims.implementing"), summary.implementing],
            [t("claims.prSubmitted"), summary.prSubmitted],
            [t("claims.merged"), summary.merged],
            [t("claims.archived"), summary.archived],
          ].map(([label, value]) => (
            <div
              className="rounded-xl border border-border bg-muted/25 p-3"
              key={label}
            >
              <dt className="text-xs text-muted-foreground">{label}</dt>
              <dd className="mt-1 font-mono text-xl font-semibold">{value}</dd>
            </div>
          ))}
        </dl>
      ) : null}

      {update.error ? <AccountRequestAlert error={update.error} /> : null}
      {remove.error ? <AccountRequestAlert error={remove.error} /> : null}
      {query.error ? <AccountRequestAlert error={query.error} /> : null}
      {moveError ? (
        <p className="text-sm text-warning" role="alert">
          {moveError}
        </p>
      ) : null}
      {update.isSuccess ? (
        <p aria-live="polite" className="text-sm text-success" role="status">
          {t("claims.updated")}
        </p>
      ) : null}

      {query.isPending ? (
        <Card>
          <CardContent
            className="p-6 text-sm text-muted-foreground"
            role="status"
          >
            {t("claims.loading")}
          </CardContent>
        </Card>
      ) : claims.length === 0 ? (
        <Card>
          <CardContent className="grid justify-items-center gap-3 p-8 text-center">
            <p className="font-semibold">{t("claims.empty")}</p>
            <p className="max-w-lg text-sm text-muted-foreground">
              {t("claims.emptyDescription")}
            </p>
          </CardContent>
        </Card>
      ) : view === "list" ? (
        <ul className="grid gap-4 lg:grid-cols-2">
          {claims.map((claim) => (
            <li key={`${claim.id}-${claim.version}`}>
              <IssueClaimCard
                busy={update.isPending || remove.isPending}
                claim={claim}
                onDelete={(id, version) => remove.mutate({ id, version })}
                onUpdate={(id, request) => update.mutate({ id, request })}
              />
            </li>
          ))}
        </ul>
      ) : (
        <div>
          <p
            className="mb-3 text-sm text-muted-foreground"
            id="kanban-instructions"
          >
            {t("claims.kanbanInstructions")}
          </p>
          <div
            aria-describedby="kanban-instructions"
            className="grid auto-cols-[minmax(18rem,1fr)] grid-flow-col gap-4 overflow-x-auto pb-3"
          >
            {statusOptions.map((status) => {
              const columnClaims = claims.filter(
                (claim) => claim.status === status,
              );
              return (
                <section
                  className="min-h-56 rounded-2xl border border-border bg-muted/20 p-3"
                  key={status}
                  onDragOver={(event) => event.preventDefault()}
                  onDrop={(event) => {
                    event.preventDefault();
                    const claim = claims.find(
                      ({ id }) => id === draggedClaimID,
                    );
                    setDraggedClaimID(null);
                    if (claim) moveClaim(claim, status);
                  }}
                >
                  <h3 className="mb-3 flex items-center justify-between gap-2 font-semibold">
                    {statusLabels[status]}
                    <Badge variant="neutral">{columnClaims.length}</Badge>
                  </h3>
                  {columnClaims.length > 0 ? (
                    <ul className="grid gap-3">
                      {columnClaims.map((claim) => (
                        <li
                          aria-label={t("claims.draggableLabel", {
                            label: `${claim.repositoryOwner}/${claim.repositoryName}#${claim.issueNumber}`,
                          })}
                          draggable={!update.isPending && !remove.isPending}
                          key={`${claim.id}-${claim.version}`}
                          onDragEnd={() => setDraggedClaimID(null)}
                          onDragStart={() => setDraggedClaimID(claim.id)}
                        >
                          <IssueClaimCard
                            busy={update.isPending || remove.isPending}
                            claim={claim}
                            onDelete={(id, version) =>
                              remove.mutate({ id, version })
                            }
                            onUpdate={(id, request) =>
                              update.mutate({ id, request })
                            }
                          />
                        </li>
                      ))}
                    </ul>
                  ) : (
                    <p className="rounded-xl border border-dashed border-border p-4 text-center text-sm text-muted-foreground">
                      {t("claims.emptyColumn")}
                    </p>
                  )}
                </section>
              );
            })}
          </div>
        </div>
      )}
    </section>
  );
}

function IssueClaimCard({
  busy,
  claim,
  onDelete,
  onUpdate,
}: {
  busy: boolean;
  claim: IssueClaim;
  onDelete: (id: string, version: number) => void;
  onUpdate: (id: string, request: IssueClaimUpdateRequest) => void;
}) {
  const { t } = useI18n();
  const statusLabels: Record<IssueClaimStatus, string> = {
    implementing: t("claims.implementing"),
    merged: t("claims.merged"),
    not_started: t("claims.notStarted"),
    pr_submitted: t("claims.prSubmitted"),
    researching: t("claims.researching"),
  };
  const [status, setStatus] = useState(claim.status);
  const [pullOwner, setPullOwner] = useState(
    claim.pullRequest?.repositoryOwner ?? claim.repositoryOwner,
  );
  const [pullRepository, setPullRepository] = useState(
    claim.pullRequest?.repositoryName ?? claim.repositoryName,
  );
  const [pullNumber, setPullNumber] = useState(
    claim.pullRequest?.number.toString() ?? "",
  );
  const needsPullRequest = status === "pr_submitted" || status === "merged";
  const parsedPullNumber = Number(pullNumber);
  const pullRequest =
    pullNumber.trim() &&
    Number.isSafeInteger(parsedPullNumber) &&
    parsedPullNumber > 0
      ? {
          number: parsedPullNumber,
          repositoryName: pullRepository.trim(),
          repositoryOwner: pullOwner.trim(),
        }
      : null;
  const label = `${claim.repositoryOwner}/${claim.repositoryName}#${claim.issueNumber}`;

  return (
    <Card className="h-full">
      <CardContent className="grid gap-4 p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="font-mono text-sm font-semibold">{label}</p>
            <div className="mt-2 flex flex-wrap gap-2">
              <Badge variant={claim.archived ? "neutral" : "accent"}>
                {claim.archived ? t("claims.archived") : t("claims.active")}
              </Badge>
              <Badge variant="neutral">
                {t("claims.issueUpstream", {
                  state: claim.observedIssueState,
                })}
              </Badge>
              {claim.pullRequest ? (
                <Badge variant="neutral">
                  {t("claims.prUpstream", { state: claim.observedPrState })}
                </Badge>
              ) : null}
            </div>
          </div>
          <ReferenceObservationButton
            request={{
              kind: "issue",
              number: claim.issueNumber,
              owner: claim.repositoryOwner,
              repositoryName: claim.repositoryName,
            }}
          />
          {claim.pullRequest ? (
            <ReferenceObservationButton
              request={{
                kind: "pull_request",
                number: claim.pullRequest.number,
                owner: claim.pullRequest.repositoryOwner,
                repositoryName: claim.pullRequest.repositoryName,
              }}
            />
          ) : null}
          <Button asChild size="small" variant="ghost">
            <Link
              to={appRoutes.issue(
                claim.repositoryOwner,
                claim.repositoryName,
                claim.issueNumber,
              )}
            >
              {t("recommendation.details")}
            </Link>
          </Button>
        </div>

        <label className="grid gap-1 text-sm font-medium">
          {t("claims.workflow")}
          <select
            className="min-h-11 rounded-xl border border-input bg-surface px-3"
            onChange={(event) =>
              setStatus(event.target.value as IssueClaimStatus)
            }
            value={status}
          >
            {statusOptions.map((option) => (
              <option key={option} value={option}>
                {statusLabels[option]}
              </option>
            ))}
          </select>
        </label>

        <fieldset className="grid gap-3 rounded-xl border border-border p-3">
          <legend className="px-1 text-sm font-semibold">
            {t("claims.linkedPr", {
              requirement: t(
                needsPullRequest ? "claims.required" : "claims.optional",
              ),
            })}
          </legend>
          <Input
            aria-label={t("claims.prOwner", { label })}
            onChange={(event) => setPullOwner(event.target.value)}
            placeholder={t("claims.owner")}
            value={pullOwner}
          />
          <Input
            aria-label={t("claims.prRepository", { label })}
            onChange={(event) => setPullRepository(event.target.value)}
            placeholder={t("claims.repository")}
            value={pullRepository}
          />
          <Input
            aria-label={t("claims.prNumberLabel", { label })}
            inputMode="numeric"
            min={1}
            onChange={(event) => setPullNumber(event.target.value)}
            placeholder={t("claims.prNumber")}
            type="number"
            value={pullNumber}
          />
          {claim.pullRequest ? (
            <a
              className="text-sm font-semibold text-accent underline-offset-4 hover:underline"
              href={externalLinks.gitHubPullRequest(
                claim.pullRequest.repositoryOwner,
                claim.pullRequest.repositoryName,
                claim.pullRequest.number,
              )}
              rel="noreferrer"
              target="_blank"
            >
              {t("claims.openPr", { number: claim.pullRequest.number })}
            </a>
          ) : null}
        </fieldset>

        <div className="flex flex-wrap gap-2">
          <Button
            disabled={
              busy ||
              (needsPullRequest && !pullRequest) ||
              Boolean(
                pullRequest &&
                (!pullRequest.repositoryOwner || !pullRequest.repositoryName),
              )
            }
            onClick={() =>
              onUpdate(claim.id, {
                archived: claim.archived,
                pullRequest,
                status,
                version: claim.version,
              })
            }
            size="small"
          >
            {t("claims.save")}
          </Button>
          <Button
            disabled={busy}
            onClick={() =>
              onUpdate(claim.id, {
                archived: !claim.archived,
                pullRequest: claim.pullRequest,
                status: claim.status,
                version: claim.version,
              })
            }
            size="small"
            variant="outline"
          >
            {claim.archived ? t("claims.restore") : t("claims.archive")}
          </Button>
          <Button
            aria-label={t("claims.deleteLabel", { label })}
            disabled={busy}
            onClick={() => onDelete(claim.id, claim.version)}
            size="small"
            variant="danger"
          >
            {t("claims.delete")}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
