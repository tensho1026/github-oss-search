import { Checkbox } from "./checkbox";

type FilterToggleProps = {
  checked: boolean;
  description: string;
  id: string;
  label: string;
  onChange: (checked: boolean) => void;
};

export function FilterToggle({
  checked,
  description,
  id,
  label,
  onChange,
}: FilterToggleProps) {
  return (
    <label
      className="flex min-w-0 cursor-pointer items-start gap-3 rounded-xl border border-border bg-muted/40 p-4 transition-colors hover:border-accent/35 hover:bg-muted"
      htmlFor={id}
    >
      <Checkbox
        checked={checked}
        id={id}
        onChange={(event) => onChange(event.target.checked)}
      />
      <span className="min-w-0 [overflow-wrap:anywhere]">
        <span className="block text-sm font-semibold">{label}</span>
        <span className="mt-1 block text-xs leading-5 text-muted-foreground">
          {description}
        </span>
      </span>
    </label>
  );
}
