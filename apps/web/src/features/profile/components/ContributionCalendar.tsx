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

const weekdays = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const levels = [
  "none",
  "first_quartile",
  "second_quartile",
  "third_quartile",
  "fourth_quartile",
] as const;

function monthLabel(date: string): string {
  return new Date(`${date}T00:00:00Z`).toUTCString().slice(8, 11);
}

function dayLabel(day: ContributionDay): string {
  return `${day.count} public ${day.count === 1 ? "contribution" : "contributions"} on ${day.date}`;
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
  if (calendar.status === "unavailable" || calendar.weeks.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Public contribution calendar</CardTitle>
          <CardDescription>
            GitHub did not provide a public daily calendar. Other profile
            evidence remains available.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  const monthLabels = calendar.weeks.map((week, index) => {
    const current = monthLabel(week.firstDay);
    const previousWeek = calendar.weeks[index - 1];
    const previous = previousWeek ? monthLabel(previousWeek.firstDay) : "";
    return current === previous ? "" : current;
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Public contribution calendar</CardTitle>
        <CardDescription>
          {calendar.total} public contributions from {calendar.from} to{" "}
          {calendar.to}. GitHub-normalized public intensity only.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div
          aria-label="Public contribution calendar. Scroll for every week."
          className="overflow-x-auto pb-3"
          tabIndex={0}
        >
          <table className="w-max border-separate border-spacing-1">
            <caption className="sr-only">
              {calendar.total} public contributions. Focus a day for details.
            </caption>
            <thead>
              <tr>
                <th className="w-9" scope="col">
                  <span className="sr-only">Weekday</span>
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
          aria-label="Contribution intensity legend from less to more"
          className="mt-3 flex items-center justify-end gap-1.5 text-xs text-muted-foreground"
        >
          <span>Less</span>
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
          <span>More</span>
        </div>
      </CardContent>
    </Card>
  );
}

function CalendarDayCell({ day }: { day: ContributionDay }) {
  const label = dayLabel(day);
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
