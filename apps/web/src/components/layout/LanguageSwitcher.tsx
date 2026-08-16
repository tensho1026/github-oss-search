import { Languages } from "lucide-react";

import { useI18n, type Locale } from "../../shared/i18n/i18n-context";
import { Icon } from "../ui/icon";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

export function LanguageSwitcher({ compact = false }: { compact?: boolean }) {
  const { locale, setLocale, t } = useI18n();

  return (
    <Select
      onValueChange={(value) => setLocale(value as Locale)}
      value={locale}
    >
      <SelectTrigger
        aria-label={t("language.label")}
        className={
          compact
            ? "min-h-9 w-9 gap-0 rounded-lg border-0 bg-transparent px-2 [&>svg:last-child]:hidden"
            : "min-h-9 gap-2 rounded-lg border-0 bg-transparent px-2"
        }
      >
        <Icon aria-hidden="true" className="size-4" icon={Languages} />
        {compact ? null : <SelectValue />}
      </SelectTrigger>
      <SelectContent align="end">
        <SelectItem value="en">{t("language.english")}</SelectItem>
        <SelectItem value="ja">{t("language.japanese")}</SelectItem>
      </SelectContent>
    </Select>
  );
}
