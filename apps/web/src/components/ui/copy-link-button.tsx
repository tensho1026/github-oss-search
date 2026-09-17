import { Check, Link2 } from "lucide-react";
import { useState } from "react";

import { useI18n } from "../../shared/i18n/i18n-context";
import { Button } from "./button";
import { Icon } from "./icon";

export function CopyLinkButton({
  href,
  label,
}: {
  href?: string;
  label?: string;
}) {
  const { t } = useI18n();
  const [status, setStatus] = useState<"copied" | "error" | "idle">("idle");

  async function copy() {
    const value = href ?? window.location.href;
    try {
      await navigator.clipboard.writeText(value);
      setStatus("copied");
    } catch {
      setStatus("error");
    }
  }

  return (
    <Button
      onClick={() => {
        void copy();
      }}
      size="small"
      type="button"
      variant="outline"
    >
      <Icon icon={status === "copied" ? Check : Link2} />
      {status === "copied"
        ? t("search.copyLinkCopied")
        : status === "error"
          ? t("search.copyLinkFailed")
          : (label ?? t("search.copyLink"))}
    </Button>
  );
}
