import { describe, expect, it } from "vitest";

import {
  gitHubUserFixture,
  profileAnalysisFixture,
} from "../../../test/profile-fixtures";
import { profileCardLines, profileMarkdown } from "./profile-export";

describe("profile export", () => {
  it("builds a portable Markdown profile from bounded public evidence", () => {
    const markdown = profileMarkdown(
      gitHubUserFixture.data,
      profileAnalysisFixture.data,
    );

    expect(markdown).toContain("# The Octocat — OSS profile");
    expect(markdown).toContain("https://github.com/octocat");
    expect(markdown).toContain("## Technologies");
    expect(markdown).toContain("## OSS Journey");
    expect(markdown).toContain("bounded public GitHub evidence");
  });

  it("limits image-card content to compact safe text lines", () => {
    const lines = profileCardLines(
      gitHubUserFixture.data,
      profileAnalysisFixture.data,
    );

    expect(lines.login).toBe("@octocat");
    expect(lines.technologies.length).toBeLessThanOrEqual(8);
    expect(lines.journey.length).toBeLessThanOrEqual(3);
  });
});
