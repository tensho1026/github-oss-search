import { Compass, Home } from "lucide-react";
import { Link } from "react-router";

import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import { Icon } from "../components/ui/icon";
import { appRoutes } from "../shared/config/app-config";
import { useI18n } from "../shared/i18n/i18n-context";

export function NotFoundPage() {
  const { t } = useI18n();

  return (
    <section className="mx-auto grid min-h-[68vh] w-full max-w-3xl content-center px-5 py-16 sm:px-8">
      <Card>
        <CardContent className="grid justify-items-start gap-5 p-8 sm:p-10">
          <span className="grid size-14 place-items-center rounded-2xl bg-accent-soft text-accent-soft-foreground">
            <Icon className="size-6" icon={Compass} />
          </span>
          <div>
            <p className="font-mono text-xs tracking-[0.16em] text-muted-foreground uppercase">
              {t("notFound.eyebrow")}
            </p>
            <h1 className="mt-3 text-3xl font-semibold tracking-[-0.045em]">
              {t("notFound.title")}
            </h1>
            <p className="mt-3 max-w-lg leading-7 text-muted-foreground">
              {t("notFound.description")}
            </p>
          </div>
          <Button asChild>
            <Link to={appRoutes.home}>
              <Icon icon={Home} />
              {t("notFound.back")}
            </Link>
          </Button>
        </CardContent>
      </Card>
    </section>
  );
}
