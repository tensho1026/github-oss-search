import { useEffect } from "react";
import { Link, useNavigate } from "react-router";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { WorkspaceDashboard } from "../features/account/components/WorkspaceDashboard";
import { useAuth } from "../features/auth/auth-context";
import { appRoutes } from "../shared/config/app-config";
import { useI18n } from "../shared/i18n/i18n-context";

export function WorkspacePage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const { markSessionExpired, query, session, signIn } = useAuth();

  useEffect(() => {
    if (query.fetchStatus === "idle" && !query.data && !query.error) {
      void query.refetch();
    }
  }, [query]);

  if (query.isFetching && !query.data) {
    return (
      <div
        aria-label={t("workspace.checking")}
        className="mx-auto grid min-h-[68vh] w-full max-w-7xl place-content-center gap-3 px-5 text-center"
        role="status"
      >
        <p className="font-semibold">{t("workspace.checking")}</p>
      </div>
    );
  }

  if (query.error) {
    return (
      <div className="mx-auto grid min-h-[68vh] w-full max-w-3xl content-center gap-5 px-5 py-12">
        <Alert variant="warning">
          <AlertTitle>{t("workspace.errorTitle")}</AlertTitle>
          <AlertDescription>{t("workspace.errorDescription")}</AlertDescription>
        </Alert>
        <Button asChild variant="outline">
          <Link to={appRoutes.search}>{t("workspace.continuePublic")}</Link>
        </Button>
      </div>
    );
  }

  if (!session?.authenticated || !session.user || !session.csrfToken) {
    return (
      <div className="mx-auto grid min-h-[68vh] w-full max-w-3xl content-center px-5 py-12">
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">
              {t("workspace.signInTitle")}
            </CardTitle>
            <CardDescription>
              {t("workspace.signInDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-3">
            {session?.configured === false ? (
              <span className="inline-flex items-center gap-2 text-sm text-muted-foreground">
                {t("workspace.notConfigured")}
              </span>
            ) : (
              <Button
                onClick={() => signIn(`${appRoutes.workspace}?tab=bookmarks`)}
              >
                {t("workspace.signInGitHub")}
              </Button>
            )}
            <Button asChild variant="outline">
              <Link to={appRoutes.search}>{t("workspace.anonymous")}</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <WorkspaceDashboard
      csrfToken={session.csrfToken}
      onAccountDeleted={async () => {
        await markSessionExpired();
        void navigate(appRoutes.home, { replace: true });
      }}
      onSessionExpired={markSessionExpired}
      user={session.user}
    />
  );
}
