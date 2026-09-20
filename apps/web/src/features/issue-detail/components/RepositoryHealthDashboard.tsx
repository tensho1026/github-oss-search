import { Badge } from "../../../components/ui/badge";
import type { IssueDetailEnvelope } from "../../../shared/api/generated";
import { useI18n } from "../../../shared/i18n/i18n-context";
import { formatDate } from "../../../shared/lib/format";
import { Section } from "./detail-section";

type Props = {
  dashboard: IssueDetailEnvelope["data"]["healthDashboard"];
};

export function RepositoryHealthDashboard({ dashboard }: Props) {
  const { locale, t } = useI18n();
  const healthCategoryLabels = {
    activity: t("detail.healthActivity"),
    beginner_friendly: t("detail.healthBeginner"),
    community: t("detail.healthCommunity"),
    security: t("detail.healthSecurity"),
  } as const;
  return (
    <Section title={t("detail.healthTitle")}>
      <p className="mb-4 text-sm text-muted-foreground">
        {t("detail.healthDescription", { version: dashboard.scoreVersion })}
      </p>
      <div className="grid gap-3 sm:grid-cols-2">
        {dashboard.categories.map((category) => (
          <details
            className="rounded-xl border border-border bg-muted/25 p-3"
            key={category.name}
          >
            <summary className="cursor-pointer font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
              {healthCategoryLabels[category.name]}{" "}
              {category.score === null
                ? t("detail.unavailable")
                : category.score}
              <Badge className="ml-2" variant="neutral">
                {t("detail.healthStatus", {
                  confidence: category.confidence,
                  status: category.status,
                })}
              </Badge>
            </summary>
            <ul className="mt-3 grid gap-2 text-xs">
              {category.components.map((component) => (
                <li key={component.key}>
                  <span className="font-semibold">
                    {t("detail.healthWeight", {
                      component: component.key.replaceAll("_", " "),
                      weight: component.weight,
                    })}
                  </span>
                  <span className="text-muted-foreground">
                    {" "}
                    ·{" "}
                    {component.score === null
                      ? t("detail.unavailable")
                      : component.score}{" "}
                    · {component.source}
                  </span>
                  <p className="text-muted-foreground">
                    {component.description}
                  </p>
                </li>
              ))}
            </ul>
            {category.warnings.map((warning) => (
              <p className="mt-2 text-xs text-warning" key={warning}>
                {warning}
              </p>
            ))}
            <p className="mt-2 text-xs text-muted-foreground">
              {t("detail.healthAnalyzed", {
                date: formatDate(category.analyzedAt, locale),
              })}
              {category.sourceVersion
                ? ` · ${t("detail.healthUpstream", {
                    version: category.sourceVersion,
                  })}`
                : ""}
            </p>
          </details>
        ))}
      </div>
      <p className="mt-4 text-xs text-muted-foreground">
        {t("detail.healthSecurityNote")}
      </p>
    </Section>
  );
}
