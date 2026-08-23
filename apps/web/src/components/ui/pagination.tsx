import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "./button";
import { Icon } from "./icon";
import { useI18n } from "../../shared/i18n/i18n-context";

type PaginationProps = {
  ariaLabel?: string;
  disabled?: boolean;
  hasNext: boolean;
  onPageChange: (page: number) => void;
  page: number;
  totalPages: number;
};

export function Pagination({
  ariaLabel,
  disabled,
  hasNext,
  onPageChange,
  page,
  totalPages,
}: PaginationProps) {
  const { t } = useI18n();
  if (totalPages < 1) {
    return null;
  }
  return (
    <nav
      aria-label={ariaLabel ?? t("pagination.defaultLabel")}
      className="grid grid-cols-2 items-center gap-3 xs:flex xs:flex-wrap xs:justify-between"
    >
      <Button
        aria-label={t("pagination.goTo", { page: Math.max(1, page - 1) })}
        disabled={disabled || page <= 1}
        onClick={() => onPageChange(page - 1)}
        className="row-start-2 w-full xs:w-auto"
        variant="outline"
      >
        <Icon icon={ChevronLeft} />
        {t("pagination.previous")}
      </Button>
      <p
        aria-live="polite"
        className="col-span-2 row-start-1 text-center text-sm text-muted-foreground xs:order-none"
      >
        {t("pagination.summary", { page, totalPages })}
      </p>
      <Button
        aria-label={t("pagination.goTo", { page: page + 1 })}
        disabled={disabled || !hasNext}
        onClick={() => onPageChange(page + 1)}
        className="row-start-2 w-full xs:w-auto"
        variant="outline"
      >
        {t("pagination.next")}
        <Icon icon={ChevronRight} />
      </Button>
    </nav>
  );
}
