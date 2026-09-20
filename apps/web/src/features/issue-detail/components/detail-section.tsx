import type { ReactNode } from "react";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "../../../components/ui/card";
import type { Evidence } from "../../../shared/api/generated";

export function Section({
  children,
  id,
  title,
}: {
  children: ReactNode;
  id?: string;
  title: string;
}) {
  return (
    <Card className="min-w-0 overflow-hidden" id={id}>
      <CardHeader className="border-b border-border bg-muted/25">
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-5 sm:p-6">{children}</CardContent>
    </Card>
  );
}

export function EvidenceList({ items }: { items: Evidence[] }) {
  return items.length > 0 ? (
    <ul className="mt-2 grid gap-1 text-xs leading-5 text-muted-foreground">
      {items.map((item) => (
        <li key={`${item.ruleId}-${item.source}-${item.description}`}>
          {item.description}
        </li>
      ))}
    </ul>
  ) : null;
}

export function Facts({ items }: { items: Array<[string, ReactNode]> }) {
  return (
    <dl className="grid grid-cols-2 gap-4">
      {items.map(([label, value]) => (
        <div key={label}>
          <dt className="text-xs text-muted-foreground">{label}</dt>
          <dd className="mt-1 font-medium">{value}</dd>
        </div>
      ))}
    </dl>
  );
}
