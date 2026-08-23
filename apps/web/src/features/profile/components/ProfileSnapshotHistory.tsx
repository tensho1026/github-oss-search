import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useRef } from "react";

import { Badge } from "../../../components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import type {
  ProfileAnalysis,
  ProfileSnapshot as MonthlyProfileSnapshot,
  ProfileSnapshotWriteRequest,
} from "../../../shared/api/generated";
import { queryKeys } from "../../../shared/query/query-keys";
import { useI18n } from "../../../shared/i18n/i18n-context";
import {
  listProfileSnapshots,
  upsertProfileSnapshot,
} from "../../account/api/account";

type Props = {
  analysis: ProfileAnalysis;
  csrfToken: string;
};

function snapshotRequest(
  analysis: ProfileAnalysis,
): ProfileSnapshotWriteRequest {
  const contribution = analysis.contributions;
  return {
    completedQuests: analysis.ossQuest.completed,
    currentStreak: analysis.contributionStreak.currentWeeks,
    frameworks: analysis.frameworks.slice(0, 20),
    languages: analysis.languages.map(({ name }) => name).slice(0, 20),
    longestStreak: analysis.contributionStreak.longestWeeks,
    mergedPullRequests: analysis.contributionPortfolio.totalMerged,
    ossActivity:
      contribution.commits.value +
      contribution.issuesOpened.value +
      contribution.pullRequestsOpened.value +
      contribution.pullRequestReviews.value,
    proficiency: analysis.proficiency
      .filter(({ level }) => level >= 1 && level <= 5)
      .slice(0, 20)
      .map(({ level, name }) => ({ level, name })),
  };
}

export function ProfileSnapshotHistory({ analysis, csrfToken }: Props) {
  const { locale, t } = useI18n();
  const queryClient = useQueryClient();
  const savedKey = useRef("");
  const query = useQuery({
    queryFn: ({ signal }) => listProfileSnapshots(signal),
    queryKey: queryKeys.account.profileSnapshots,
  });
  const mutation = useMutation({
    mutationFn: (request: ProfileSnapshotWriteRequest) =>
      upsertProfileSnapshot(request, csrfToken),
    async onSuccess() {
      await queryClient.invalidateQueries({
        queryKey: queryKeys.account.profileSnapshots,
      });
    },
  });
  const request = useMemo(() => snapshotRequest(analysis), [analysis]);
  const requestKey = JSON.stringify(request);
  useEffect(() => {
    if (savedKey.current === requestKey) return;
    savedKey.current = requestKey;
    mutation.mutate(request);
  }, [mutation, request, requestKey]);

  const items = query.data?.data.items ?? [];
  const latest = items.at(-1);
  const previous = items.at(-2);
  const changes = latest ? snapshotChanges(latest, previous) : undefined;
  const formatMonth = (value: string) =>
    new Intl.DateTimeFormat(locale, { month: "short", year: "numeric" }).format(
      new Date(value),
    );

  return (
    <Card className="mt-5 border-accent/20">
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <CardTitle>{t("profile.snapshotTitle")}</CardTitle>
            <CardDescription>
              {t("profile.snapshotDescription")}
            </CardDescription>
          </div>
          <Badge variant={mutation.isError ? "warning" : "success"}>
            {mutation.isPending
              ? t("profile.snapshotSaving")
              : mutation.isError
                ? t("profile.snapshotUnavailable")
                : t("profile.snapshotPrivate")}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        {query.isPending ? (
          <p className="text-sm text-muted-foreground" role="status">
            {t("profile.snapshotLoading")}
          </p>
        ) : latest && changes ? (
          <div className="grid gap-5">
            <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <SnapshotMetric
                label={t("profile.snapshotActivity")}
                value={latest.ossActivity}
                delta={changes.activity}
              />
              <SnapshotMetric
                label={t("profile.snapshotMerged")}
                value={latest.mergedPullRequests}
                delta={changes.merged}
              />
              <SnapshotMetric
                label={t("profile.snapshotQuest")}
                value={latest.completedQuests}
                delta={changes.quests}
              />
              <SnapshotMetric
                label={t("profile.snapshotStreak")}
                value={latest.currentStreak}
                suffix={t("profile.snapshotWeeks")}
              />
            </dl>
            <div className="grid gap-4 lg:grid-cols-2">
              <ChangeList
                empty={t("profile.snapshotNoNewTechnology")}
                label={t("profile.snapshotNewTechnology")}
                values={changes.newTechnologies}
              />
              <ChangeList
                empty={t("profile.snapshotNoProficiencyChange")}
                label={t("profile.snapshotProficiencyChange")}
                values={changes.proficiency}
              />
            </div>
            <ol className="flex gap-2 overflow-x-auto pb-1">
              {items.map((item) => (
                <li
                  className="min-w-28 rounded-xl border border-border bg-muted/30 p-3 text-center"
                  key={item.month}
                >
                  <p className="text-xs text-muted-foreground">
                    {formatMonth(item.month)}
                  </p>
                  <p className="mt-1 font-mono font-semibold">
                    {item.mergedPullRequests} PR
                  </p>
                </li>
              ))}
            </ol>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            {t("profile.snapshotFirstMonth")}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function snapshotChanges(
  latest: MonthlyProfileSnapshot,
  previous?: MonthlyProfileSnapshot,
) {
  const before = previous ?? {
    completedQuests: 0,
    frameworks: [],
    languages: [],
    mergedPullRequests: 0,
    ossActivity: 0,
    proficiency: [],
  };
  const priorTechnologies = new Set(
    [...before.languages, ...before.frameworks].map((value) =>
      value.toLowerCase(),
    ),
  );
  const priorLevels = new Map(
    before.proficiency.map(({ level, name }) => [name.toLowerCase(), level]),
  );
  return {
    activity: latest.ossActivity - before.ossActivity,
    merged: latest.mergedPullRequests - before.mergedPullRequests,
    newTechnologies: [...latest.languages, ...latest.frameworks].filter(
      (value) => !priorTechnologies.has(value.toLowerCase()),
    ),
    proficiency: latest.proficiency
      .map(({ level, name }) => ({
        delta: level - (priorLevels.get(name.toLowerCase()) ?? level),
        name,
      }))
      .filter(({ delta }) => delta !== 0)
      .map(({ delta, name }) => `${name} ${delta > 0 ? "+" : ""}${delta}`),
    quests: latest.completedQuests - before.completedQuests,
  };
}

function SnapshotMetric({
  label,
  value,
  delta,
  suffix = "",
}: {
  label: string;
  value: number;
  delta?: number;
  suffix?: string;
}) {
  return (
    <div className="rounded-xl bg-muted p-4">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-2 font-mono text-xl font-semibold">
        {value}
        {suffix ? ` ${suffix}` : ""}
        {delta && delta !== 0 ? (
          <span className="ml-2 text-xs text-accent">
            {delta > 0 ? "+" : ""}
            {delta}
          </span>
        ) : null}
      </dd>
    </div>
  );
}

function ChangeList({
  empty,
  label,
  values,
}: {
  empty: string;
  label: string;
  values: readonly string[];
}) {
  return (
    <section className="rounded-xl border border-border p-4">
      <h3 className="text-sm font-semibold">{label}</h3>
      <div className="mt-2 flex flex-wrap gap-2">
        {values.length > 0 ? (
          values.map((value) => (
            <Badge key={value} variant="accent">
              {value}
            </Badge>
          ))
        ) : (
          <span className="text-sm text-muted-foreground">{empty}</span>
        )}
      </div>
    </section>
  );
}
