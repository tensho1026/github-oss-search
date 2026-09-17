import type { CompareReference } from "../../issue-compare/model/compare-location";
import { maximumComparedIssues } from "../../issue-compare/model/compare-location";

export const searchSelectionParameterName = "select";

const referencePattern = /^([^/]{1,39})\/([^#]{1,100})#([1-9]\d*)$/;

export function searchSelectionKey(reference: CompareReference): string {
  return `${reference.owner.toLowerCase()}/${reference.repository.toLowerCase()}#${reference.issueNumber}`;
}

export function decodeSearchSelection(
  parameters: URLSearchParams,
): CompareReference[] {
  const selected: CompareReference[] = [];
  const seen = new Set<string>();
  for (const raw of parameters.getAll(searchSelectionParameterName)) {
    const match = referencePattern.exec(raw.trim());
    if (!match) {
      continue;
    }
    const issueNumber = Number(match[3]);
    if (!Number.isSafeInteger(issueNumber)) {
      continue;
    }
    const reference = {
      issueNumber,
      owner: match[1]!,
      repository: match[2]!,
    };
    const key = searchSelectionKey(reference);
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    selected.push(reference);
    if (selected.length === maximumComparedIssues) {
      break;
    }
  }
  return selected;
}

export function appendSearchSelection(
  parameters: URLSearchParams,
  selected: readonly CompareReference[],
): void {
  const seen = new Set<string>();
  for (const reference of selected) {
    const key = searchSelectionKey(reference);
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    parameters.append(
      searchSelectionParameterName,
      `${reference.owner}/${reference.repository}#${reference.issueNumber}`,
    );
    if (seen.size === maximumComparedIssues) {
      break;
    }
  }
}

export function toggleSearchSelection(
  selected: readonly CompareReference[],
  reference: CompareReference,
): CompareReference[] {
  const key = searchSelectionKey(reference);
  const exists = selected.some((item) => searchSelectionKey(item) === key);
  if (exists) {
    return selected.filter((item) => searchSelectionKey(item) !== key);
  }
  if (selected.length >= maximumComparedIssues) {
    return [...selected];
  }
  return [...selected, reference];
}
