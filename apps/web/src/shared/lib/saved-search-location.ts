export const savedSearchParameterName = "saved";

const savedSearchIdPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export function decodeSavedSearchId(
  parameters: URLSearchParams,
): string | undefined {
  const values = parameters.getAll(savedSearchParameterName);
  if (values.length !== 1) {
    return undefined;
  }
  const value = values[0]?.trim() ?? "";
  return savedSearchIdPattern.test(value) ? value : undefined;
}

export function appendSavedSearchId(
  parameters: URLSearchParams,
  savedSearchId: string | undefined,
): void {
  if (!savedSearchId || !savedSearchIdPattern.test(savedSearchId)) {
    return;
  }
  parameters.set(savedSearchParameterName, savedSearchId);
}
