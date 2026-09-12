import { describe, expect, it } from "vitest";

import { ApiError } from "../../../shared/api/client";
import {
  prioritizedProfileError,
  profileErrorPresentation,
} from "./profile-error";

describe("profile error presentation", () => {
  it("explains invalid usernames, unexpected API failures, and network errors", () => {
    const invalidUsername = profileErrorPresentation(
      new ApiError({
        code: "INVALID_USERNAME",
        message: "invalid username",
        requestId: "req_400",
        status: 400,
      }),
      "bad/name",
    );
    expect(invalidUsername).toMatchObject({
      requestId: "req_400",
      retryable: false,
      title: "This username needs a second look",
      tone: "warning",
    });

    const unexpectedApiFailure = profileErrorPresentation(
      new ApiError({
        code: "GITHUB_API_ERROR",
        message: "upstream",
        status: 502,
      }),
      "octocat",
    );
    expect(unexpectedApiFailure).toMatchObject({
      retryable: true,
      title: "The signal dropped",
      tone: "danger",
    });

    expect(
      profileErrorPresentation(new Error("offline"), "octocat"),
    ).toMatchObject({
      retryable: true,
      title: "Connection interrupted",
      tone: "danger",
    });
  });

  it("makes not-found and rate-limit failures recoverable", () => {
    const notFound = profileErrorPresentation(
      new ApiError({
        code: "GITHUB_USER_NOT_FOUND",
        message: "not found",
        requestId: "req_404",
        status: 404,
      }),
      "missing",
    );
    expect(notFound).toMatchObject({
      requestId: "req_404",
      retryable: false,
      title: "Profile not found",
    });

    const rateLimit = profileErrorPresentation(
      new ApiError({
        code: "GITHUB_RATE_LIMIT_EXCEEDED",
        message: "limited",
        status: 429,
      }),
      "octocat",
    );
    expect(rateLimit).toMatchObject({
      retryable: true,
      title: "GitHub needs a breather",
    });
  });

  it("prioritizes actionable rate limits across parallel requests", () => {
    const upstream = new ApiError({
      code: "GITHUB_API_ERROR",
      message: "upstream",
      status: 502,
    });
    const limited = new ApiError({
      code: "GITHUB_RATE_LIMIT_EXCEEDED",
      message: "limited",
      status: 429,
    });
    expect(prioritizedProfileError([upstream, limited])).toBe(limited);
  });

  it("falls back to not-found, the first error, or no error", () => {
    const notFound = new ApiError({
      code: "GITHUB_USER_NOT_FOUND",
      message: "not found",
      status: 404,
    });
    const upstream = new ApiError({
      code: "GITHUB_API_ERROR",
      message: "upstream",
      status: 502,
    });

    expect(prioritizedProfileError([upstream, notFound])).toBe(notFound);
    expect(prioritizedProfileError([null, upstream])).toBe(upstream);
    expect(prioritizedProfileError([null])).toBeNull();
  });
});
