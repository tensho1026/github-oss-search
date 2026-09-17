import { X } from "lucide-react";

import { useI18n } from "../../shared/i18n/i18n-context";
import { Badge } from "./badge";
import { Icon } from "./icon";

export function FilterChipList({
  chips,
}: {
  chips: ReadonlyArray<{ id: string; label: string; onRemove: () => void }>;
}) {
  const { t } = useI18n();
  if (chips.length === 0) {
    return null;
  }
  return (
    <ul aria-label={t("search.activeFilters")} className="flex flex-wrap gap-2">
      {chips.map((chip) => (
        <li key={chip.id}>
          <Badge className="pr-1" variant="accent">
            {chip.label}
            <button
              aria-label={t("search.removeFilter", { label: chip.label })}
              className="inline-flex size-6 items-center justify-center rounded-full outline-none hover:bg-accent/15 focus-visible:ring-2 focus-visible:ring-ring"
              onClick={chip.onRemove}
              type="button"
            >
              <Icon className="size-3" icon={X} />
            </button>
          </Badge>
        </li>
      ))}
    </ul>
  );
}
