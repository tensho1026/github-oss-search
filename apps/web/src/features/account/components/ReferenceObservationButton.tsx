import { useMutation } from "@tanstack/react-query";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import type { GitHubReferenceObservationRequest } from "../../../shared/api/generated";
import { observeGitHubReference } from "../../../shared/api/references";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { AccountRequestAlert } from "./AccountRequestAlert";

export function ReferenceObservationButton({
  request,
}: {
  request: GitHubReferenceObservationRequest;
}) {
  const { t } = useI18n();
  const observation = useMutation({
    mutationFn: () => observeGitHubReference(request),
  });
  return (
    <>
      <Button
        disabled={observation.isPending}
        onClick={() => observation.mutate()}
        size="small"
        variant="outline"
      >
        {observation.isPending ? t("reference.checking") : t("reference.check")}
      </Button>
      {observation.data ? (
        <Badge
          variant={
            observation.data.data.state === "available" ||
            observation.data.data.state === "open" ||
            observation.data.data.state === "merged"
              ? "success"
              : "warning"
          }
        >
          {t("reference.current", { state: observation.data.data.state })}
        </Badge>
      ) : null}
      {observation.error ? (
        <AccountRequestAlert error={observation.error} />
      ) : null}
    </>
  );
}
