export function formatCompactNumber(value: number, locale = "en"): string {
  return new Intl.NumberFormat(locale, {
    maximumFractionDigits: 1,
    notation: "compact",
  }).format(Math.max(0, value));
}

export function formatDate(value: string, locale = "en-GB"): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "Unknown"
    : new Intl.DateTimeFormat(locale, {
        day: "numeric",
        month: "short",
        year: "numeric",
      }).format(date);
}

export function formatDuration(seconds: number | null, locale = "en"): string {
  if (seconds === null || !Number.isFinite(seconds) || seconds < 0) {
    return locale === "ja" ? "利用不可" : "Unavailable";
  }
  if (seconds < 60) {
    return locale === "ja" ? "1分未満" : "< 1 min";
  }
  if (seconds < 3600) {
    const minutes = Math.round(seconds / 60);
    return locale === "ja" ? `${minutes}分` : `${minutes} min`;
  }
  if (seconds < 172_800) {
    const hours = Math.round(seconds / 3600);
    return locale === "ja"
      ? `${hours}時間`
      : `${hours} ${hours === 1 ? "hr" : "hrs"}`;
  }
  const days = Math.round(seconds / 86_400);
  return locale === "ja" ? `${days}日` : `${days} days`;
}

export function formatPercentage(value: number): string {
  return `${Math.max(0, Math.min(100, Math.round(value)))}%`;
}

export function formatRating(value: number, maximum = 5): string {
  const boundedMaximum = Math.max(1, Math.round(maximum));
  const boundedValue = Math.max(0, Math.min(boundedMaximum, Math.round(value)));
  return `${"★".repeat(boundedValue)}${"☆".repeat(boundedMaximum - boundedValue)}`;
}
