import type {
  ContributionCalendar as ContributionCalendarData,
  ContributionDay,
} from "../../../shared/api/generated";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import { cn } from "../../../shared/lib/cn";
import { useI18n, type Locale } from "../../../shared/i18n/i18n-context";
import { formatCompactNumber, formatDate } from "../../../shared/lib/format";

const levels = [
  "none",
  "first_quartile",
  "second_quartile",
  "third_quartile",
  "fourth_quartile",
] as const;

function monthLabel(date: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale === "ja" ? "ja-JP" : "en-US", {
    month: "short",
    timeZone: "UTC",
  }).format(new Date(`${date}T00:00:00Z`));
}

const levelClasses: Record<ContributionDay["level"], string> = {
  none: "border-border/75 bg-muted/45",
  first_quartile: "border-accent/20 bg-accent-soft",
  second_quartile: "border-accent/25 bg-accent/20",
  third_quartile: "border-accent/35 bg-accent opacity-85",
  fourth_quartile: "border-accent bg-accent",
};

export function ContributionCalendar({
  calendar,
}: {
  calendar: ContributionCalendarData;
}) {
  const { locale, t } = useI18n();
  const weekdays = Array.from({ length: 7 }, (_, weekday) =>
    new Intl.DateTimeFormat(locale === "ja" ? "ja-JP" : "en-US", {
      timeZone: "UTC",
      weekday: "short",
    }).format(new Date(Date.UTC(2026, 7, 16 + weekday))),
  );
  if (calendar.status === "unavailable" || calendar.weeks.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{t("calendar.title")}</CardTitle>
          <CardDescription>{t("calendar.unavailable")}</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  const monthLabels = calendar.weeks.map((week, index) => {
    const current = monthLabel(week.firstDay, locale);
    const previousWeek = calendar.weeks[index - 1];
    const previous = previousWeek
      ? monthLabel(previousWeek.firstDay, locale)
      : "";
    return current === previous ? "" : current;
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("calendar.title")}</CardTitle>
        <CardDescription>
          {t("calendar.summary", {
            from: calendar.from
              ? formatDate(calendar.from, locale)
              : t("detail.unknown"),
            to: calendar.to
              ? formatDate(calendar.to, locale)
              : t("detail.unknown"),
            total: formatCompactNumber(calendar.total, locale),
          })}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div
          aria-label={t("calendar.scrollLabel")}
          className="overflow-x-auto pb-3"
          tabIndex={0}
        >
          <table className="w-max border-separate border-spacing-1">
            <caption className="sr-only">
              {t("calendar.caption", {
                total: formatCompactNumber(calendar.total, locale),
              })}
            </caption>
            <thead>
              <tr>
                <th className="w-9" scope="col">
                  <span className="sr-only">{t("calendar.weekday")}</span>
                </th>
                {calendar.weeks.map((week, index) => (
                  <th
                    className="h-5 min-w-3 whitespace-nowrap text-left text-xs font-medium text-muted-foreground"
                    key={week.firstDay}
                    scope="col"
                  >
                    {monthLabels[index]}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {weekdays.map((weekday, weekdayIndex) => (
                <tr key={weekday}>
                  <th
                    className="pr-1 text-right text-xs font-normal text-muted-foreground"
                    scope="row"
                  >
                    {weekday}
                  </th>
                  {calendar.weeks.map((week) => {
                    const day = week.days.find(
                      (candidate) => candidate.weekday === weekdayIndex,
                    );
                    return (
                      <td className="p-0" key={`${week.index}-${weekday}`}>
                        {day ? <CalendarDayCell day={day} /> : null}
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div
          aria-label={t("calendar.legend")}
          className="mt-3 flex items-center justify-end gap-1.5 text-xs text-muted-foreground"
        >
          <span>{t("calendar.less")}</span>
          {levels.map((level) => (
            <span
              aria-hidden="true"
              className={cn(
                "size-3 rounded-[0.2rem] border",
                levelClasses[level],
              )}
              key={level}
            />
          ))}
          <span>{t("calendar.more")}</span>
        </div>
      </CardContent>
    </Card>
  );
}

function CalendarDayCell({ day }: { day: ContributionDay }) {
  const { locale, t } = useI18n();
  const label = t("calendar.dayLabel", {
    count: formatCompactNumber(day.count, locale),
    date: formatDate(day.date, locale),
  });
  return (
    <span
      aria-label={label}
      className={cn(
        "block size-3 rounded-sm border outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1",
        levelClasses[day.level],
      )}
      role="img"
      tabIndex={0}
      title={label}
    />
  );
}
