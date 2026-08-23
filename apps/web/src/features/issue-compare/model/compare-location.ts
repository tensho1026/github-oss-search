export type CompareReference = {
  issueNumber: number;
  owner: string;
  repository: string;
};

export type CompareLocation = {
  message?: string;
  references: CompareReference[];
  returnTo: string;
  skills: string[];
  valid: boolean;
};

export const maximumComparedIssues = 3;
const maximumSkills = 20;
const referencePattern = /^([^/]{1,39})\/([^#]{1,100})#([1-9]\d*)$/;

export function encodeCompareLocation(
  references: readonly CompareReference[],
  skills: readonly string[],
  returnTo: string,
): URLSearchParams {
  const parameters = new URLSearchParams();
  for (const reference of references.slice(0, maximumComparedIssues)) {
    parameters.append(
      "issue",
      `${reference.owner}/${reference.repository}#${reference.issueNumber}`,
    );
  }
  for (const skill of skills.slice(0, maximumSkills)) {
    const normalized = skill.trim();
    if (normalized) parameters.append("skills", normalized);
  }
  if (returnTo.startsWith("/") && !returnTo.startsWith("//")) {
    parameters.set("returnTo", returnTo.slice(0, 2048));
  }
  return parameters;
}

export function decodeCompareLocation(
  parameters: URLSearchParams,
): CompareLocation {
  const rawReferences = parameters.getAll("issue");
  if (
    rawReferences.length < 2 ||
    rawReferences.length > maximumComparedIssues
  ) {
    return invalid("Select two or three issues to compare.");
  }
  const references: CompareReference[] = [];
  const seen = new Set<string>();
  for (const raw of rawReferences) {
    const match = referencePattern.exec(raw);
    if (!match) return invalid("One of the issue references is invalid.");
    const issueNumber = Number(match[3]);
    if (!Number.isSafeInteger(issueNumber)) {
      return invalid("One of the issue references is invalid.");
    }
    const reference = {
      owner: match[1]!,
      repository: match[2]!,
      issueNumber,
    };
    const key = `${reference.owner.toLowerCase()}/${reference.repository.toLowerCase()}#${issueNumber}`;
    if (seen.has(key)) return invalid("Each compared issue must be unique.");
    seen.add(key);
    references.push(reference);
  }
  const skills = parameters
    .getAll("skills")
    .map((skill) => skill.trim())
    .filter(Boolean);
  if (
    skills.length > maximumSkills ||
    skills.some((skill) => skill.length > 64)
  ) {
    return invalid("The comparison contains too many or invalid skills.");
  }
  const rawReturnTo = parameters.get("returnTo") ?? "/search";
  const returnTo =
    rawReturnTo.startsWith("/") &&
    !rawReturnTo.startsWith("//") &&
    rawReturnTo.length <= 2048
      ? rawReturnTo
      : "/search";
  return { references, returnTo, skills, valid: true };
}

function invalid(message: string): CompareLocation {
  return {
    message,
    references: [],
    returnTo: "/search",
    skills: [],
    valid: false,
  };
}
