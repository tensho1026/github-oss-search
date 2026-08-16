import { useEffect, useState } from "react";
import { useSearchParams } from "react-router";

import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "../../../components/ui/alert";
import { useAuth } from "../auth-context";
import { useI18n } from "../../../shared/i18n/i18n-context";

type AuthMarker = "denied" | "error" | "success";

function readMarker(parameters: URLSearchParams): AuthMarker | undefined {
  const values = parameters.getAll("auth");
  if (values.length !== 1) {
    return undefined;
  }
  const value = values[0];
  return value === "denied" || value === "error" || value === "success"
    ? value
    : undefined;
}

export function AuthFeedback() {
  const { t } = useI18n();
  const [parameters, setParameters] = useSearchParams();
  const { query } = useAuth();
  const { refetch } = query;
  const [marker] = useState(() => readMarker(parameters));

  useEffect(() => {
    if (!parameters.has("auth")) {
      return;
    }
    const next = new URLSearchParams(parameters);
    next.delete("auth");
    setParameters(next, { replace: true });
  }, [parameters, setParameters]);

  useEffect(() => {
    if (marker === "success") {
      void refetch();
    }
  }, [marker, refetch]);

  if (!marker) {
    return null;
  }

  const content = {
    denied: {
      description: t("auth.deniedDescription"),
      title: t("auth.deniedTitle"),
      variant: "info" as const,
    },
    error: {
      description: t("auth.errorDescription"),
      title: t("auth.errorTitle"),
      variant: "danger" as const,
    },
    success: {
      description: t("auth.successDescription"),
      title: t("auth.successTitle"),
      variant: "success" as const,
    },
  }[marker];

  return (
    <div className="mx-auto w-full max-w-7xl px-5 pt-4 sm:px-8 lg:px-10">
      <Alert variant={content.variant}>
        <AlertTitle>{content.title}</AlertTitle>
        <AlertDescription>{content.description}</AlertDescription>
      </Alert>
    </div>
  );
}
