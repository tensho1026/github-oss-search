import { GitCompareArrows, X } from "lucide-react";
import { Link } from "react-router";

import { Button } from "../../../components/ui/button";
import { Icon } from "../../../components/ui/icon";
import { appRoutes } from "../../../shared/config/app-config";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { AddIssueTasksButton } from "../../account/components/AddIssueTasksButton";
import {
  encodeCompareLocation,
  type CompareReference,
} from "../../issue-compare/model/compare-location";

export function IssueSelectionToolbar({
  onClear,
  references,
  returnTo,
  skills,
}: {
  onClear: () => void;
  references: readonly CompareReference[];
  returnTo: string;
  skills: readonly string[];
}) {
  const { t } = useI18n();
  const compareParameters = encodeCompareLocation(references, skills, returnTo);
  const comparePath = `${appRoutes.compare}?${compareParameters.toString()}`;

  if (references.length === 0) return null;
  return (
    <div className="sticky bottom-4 z-20 rounded-2xl border border-accent/40 bg-surface/95 p-4 shadow-lg backdrop-blur">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <strong>{t("selection.count", { count: references.length })}</strong>
        <div className="flex flex-wrap gap-2">
          <Button
            asChild={references.length >= 2}
            disabled={references.length < 2}
            size="small"
          >
            {references.length >= 2 ? (
              <Link to={comparePath}>
                <Icon icon={GitCompareArrows} />
                {t("selection.compare")}
              </Link>
            ) : (
              <span>
                <Icon icon={GitCompareArrows} />
                {t("selection.compare")}
              </span>
            )}
          </Button>
          <AddIssueTasksButton references={references} returnTo={returnTo} />
          <Button
            aria-label={t("selection.clear")}
            onClick={onClear}
            size="small"
            variant="ghost"
          >
            <Icon icon={X} />
          </Button>
        </div>
      </div>
      {references.length < 2 ? (
        <p className="mt-2 text-xs text-muted-foreground">
          {t("selection.compareHint")}
        </p>
      ) : null}
    </div>
  );
}
