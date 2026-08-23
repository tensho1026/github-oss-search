package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/repository"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

const (
	maxRepositorySearchQueryBytes        = 1024
	maxRepositorySearchResponseBytes     = 8 << 20
	maxRepositoryEnrichmentQueryBytes    = 64 << 10
	maxRepositoryEnrichmentResponseBytes = 8 << 20
	maxRepositoryTopics                  = 20
)

// SearchRepositories retrieves exactly one bounded REST search window.
// Expensive issue-label and documentation evidence is intentionally absent and
// loaded in one GraphQL request only for the shortlist.
func (c *Client) SearchRepositories(
	ctx context.Context,
	criteria repository.DiscoveryCriteria,
	limit int,
) (port.GitHubRepositoryDiscoveryResult, error) {
	if limit < 1 || limit > repository.MaximumDiscoveryCandidateResults {
		return port.GitHubRepositoryDiscoveryResult{}, fmt.Errorf(
			"GitHub repository search limit must be between 1 and %d",
			repository.MaximumDiscoveryCandidateResults,
		)
	}
	searchQuery, err := buildRepositorySearchQuery(criteria, c.now())
	if err != nil {
		return port.GitHubRepositoryDiscoveryResult{}, err
	}
	endpoint := *c.baseURL
	endpoint.Path = path.Join(endpoint.Path, "search", "repositories")
	query := endpoint.Query()
	query.Set("q", searchQuery)
	query.Set("per_page", strconv.Itoa(limit))
	query.Set("page", "1")
	query.Set("sort", "updated")
	query.Set("order", "desc")
	endpoint.RawQuery = query.Encode()

	response, err := c.do(
		ctx,
		operationSearchRepositories,
		endpoint.String(),
	)
	if err != nil {
		return port.GitHubRepositoryDiscoveryResult{}, err
	}
	defer response.Body.Close()
	headerRateLimit := parseRateLimit(response.Header)
	if statusErr := responseError(response.StatusCode, headerRateLimit); statusErr != nil {
		return port.GitHubRepositoryDiscoveryResult{}, statusErr
	}

	var payload restRepositoryDiscoveryEnvelope
	if decodeErr := decodeBoundedJSON(
		response.Body,
		maxRepositorySearchResponseBytes,
		"GitHub REST repository search response",
		&payload,
	); decodeErr != nil {
		return port.GitHubRepositoryDiscoveryResult{}, decodeErr
	}
	result, err := normalizeRESTRepositorySearch(
		payload,
		limit,
		headerRateLimit,
	)
	if err != nil {
		return port.GitHubRepositoryDiscoveryResult{}, err
	}
	c.logger.Debug(
		"GitHub REST repository search response received",
		"status", response.StatusCode,
		"candidateCount", len(result.Candidates),
		"upstreamTotal", result.TotalCount,
		"incompleteResults", result.IncompleteResults,
		"rateLimitKnown", result.RateLimit.Known,
		"rateLimitRemaining", result.RateLimit.Remaining,
	)
	return result, nil
}

// EnrichRepositories loads README and contribution-file evidence in one
// dynamically bounded GraphQL request. It never executes repository code.
func (c *Client) EnrichRepositories(
	ctx context.Context,
	repositories []repository.Summary,
) (port.GitHubRepositoryEnrichmentResult, error) {
	if len(repositories) == 0 {
		return port.GitHubRepositoryEnrichmentResult{
			Items: map[string]repository.DiscoveryEnrichment{},
		}, nil
	}
	if len(repositories) > repository.MaximumDiscoveryEnrichmentResults {
		return port.GitHubRepositoryEnrichmentResult{}, fmt.Errorf(
			"GitHub repository enrichment cannot contain more than %d repositories",
			repository.MaximumDiscoveryEnrichmentResults,
		)
	}
	query, aliases, unavailable, err := buildRepositoryEnrichmentQuery(
		repositories,
	)
	if err != nil {
		return port.GitHubRepositoryEnrichmentResult{}, err
	}
	if len(aliases) == 0 {
		return port.GitHubRepositoryEnrichmentResult{
			Items:             unavailable,
			IncompleteResults: true,
		}, nil
	}
	requestPayload, err := json.Marshal(struct {
		Query string `json:"query"`
	}{Query: query})
	if err != nil {
		return port.GitHubRepositoryEnrichmentResult{}, upstreamDecodeError(
			"GitHub GraphQL repository enrichment request",
			err,
		)
	}

	response, err := c.graphQLRequest(ctx, requestPayload)
	if err != nil {
		return port.GitHubRepositoryEnrichmentResult{}, err
	}
	defer response.Body.Close()
	headerRateLimit := parseRateLimit(response.Header)
	if statusErr := responseError(response.StatusCode, headerRateLimit); statusErr != nil {
		return port.GitHubRepositoryEnrichmentResult{}, statusErr
	}

	var payload graphQLRepositoryEnrichmentEnvelope
	if decodeErr := decodeBoundedJSON(
		response.Body,
		maxRepositoryEnrichmentResponseBytes,
		"GitHub GraphQL repository enrichment response",
		&payload,
	); decodeErr != nil {
		return port.GitHubRepositoryEnrichmentResult{}, decodeErr
	}
	rateLimit, err := enrichmentRateLimit(payload.Data, headerRateLimit)
	if err != nil {
		return port.GitHubRepositoryEnrichmentResult{}, err
	}
	if payload.Data == nil {
		if len(payload.Errors) > 0 {
			return port.GitHubRepositoryEnrichmentResult{},
				graphQLRequestError(
					"GitHub GraphQL repository enrichment",
					payload.Errors,
					rateLimit,
				)
		}
		return port.GitHubRepositoryEnrichmentResult{}, upstreamDecodeError(
			"GitHub GraphQL repository enrichment response",
			errors.New("does not contain data"),
		)
	}

	items := unavailable
	incomplete := len(unavailable) > 0 || len(payload.Errors) > 0
	for alias, fullName := range aliases {
		raw, present := payload.Data[alias]
		if !present || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			items[fullName] = repository.DiscoveryEnrichment{}
			incomplete = true
			continue
		}
		var node graphQLRepositoryEnrichmentNode
		if err := json.Unmarshal(raw, &node); err != nil {
			items[fullName] = repository.DiscoveryEnrichment{}
			incomplete = true
			continue
		}
		enrichment, err := node.toDomain(fullName)
		if err != nil {
			items[fullName] = repository.DiscoveryEnrichment{}
			incomplete = true
			continue
		}
		items[fullName] = enrichment
	}
	return port.GitHubRepositoryEnrichmentResult{
		Items:             items,
		IncompleteResults: incomplete,
		RateLimit:         rateLimit,
	}, nil
}

func (c *Client) graphQLRequest(
	ctx context.Context,
	payload []byte,
) (*http.Response, error) {
	endpoint := *c.baseURL
	endpoint.Path = path.Join(endpoint.Path, "graphql")
	endpoint.RawQuery = ""
	return c.doRequest(ctx, operationEnrichRepositories, func() (*http.Request, error) {
		request, err := c.newRequest(
			ctx,
			http.MethodPost,
			endpoint.String(),
			bytes.NewReader(payload),
		)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-Type", "application/json")
		return request, nil
	})
}

func buildRepositorySearchQuery(
	criteria repository.DiscoveryCriteria,
	now time.Time,
) (string, error) {
	parts := []string{"is:public"}
	if criteria.ExcludesArchived() {
		parts = append(parts, "archived:false")
	}
	switch criteria.ForkPolicy() {
	case repository.ForkPolicyInclude:
		parts = append(parts, "fork:true")
	case repository.ForkPolicyOnly:
		parts = append(parts, "fork:only")
	}
	parts = append(
		parts,
		"stars:>="+strconv.Itoa(criteria.MinimumStars()),
		"forks:>="+strconv.Itoa(criteria.MinimumForks()),
		"pushed:>="+now.UTC().
			AddDate(0, 0, -criteria.UpdatedWithinDays()).
			Format(time.DateOnly),
	)
	if languages := criteria.Languages(); len(languages) > 0 {
		for _, language := range languages {
			parts = append(
				parts,
				`language:"`+language.String()+`"`,
			)
		}
	}
	if licenses := criteria.Licenses(); len(licenses) > 0 {
		for _, license := range licenses {
			parts = append(
				parts,
				"license:"+strings.ToLower(license.String()),
			)
		}
	}
	query := strings.Join(parts, " ")
	if len(query) > maxRepositorySearchQueryBytes {
		return "", fmt.Errorf(
			"%w: the encoded GitHub repository query exceeds %d bytes",
			repository.ErrInvalidDiscoveryCriteria,
			maxRepositorySearchQueryBytes,
		)
	}
	return query, nil
}

func normalizeRESTRepositorySearch(
	payload restRepositoryDiscoveryEnvelope,
	limit int,
	rateLimit port.RateLimit,
) (port.GitHubRepositoryDiscoveryResult, error) {
	if payload.TotalCount < 0 ||
		payload.TotalCount < len(payload.Items) ||
		len(payload.Items) > limit {
		return port.GitHubRepositoryDiscoveryResult{}, upstreamDecodeError(
			"GitHub REST repository search response",
			errors.New("contains invalid result counts"),
		)
	}

	candidates := make(
		[]repository.DiscoveryCandidate,
		0,
		len(payload.Items),
	)
	for index, item := range payload.Items {
		candidate, mappingErr := item.toDomain()
		if mappingErr != nil {
			return port.GitHubRepositoryDiscoveryResult{}, upstreamDecodeError(
				"GitHub REST repository search response",
				fmt.Errorf("item %d: %w", index, mappingErr),
			)
		}
		candidates = append(candidates, candidate)
	}
	return port.GitHubRepositoryDiscoveryResult{
		Candidates: candidates,
		TotalCount: payload.TotalCount,
		IncompleteResults: payload.IncompleteResults ||
			payload.TotalCount > len(payload.Items),
		RateLimit: rateLimit,
	}, nil
}

func buildRepositoryEnrichmentQuery(
	repositories []repository.Summary,
) (
	string,
	map[string]string,
	map[string]repository.DiscoveryEnrichment,
	error,
) {
	var query strings.Builder
	query.WriteString("query IssueScoutRepositoryEnrichment {\n")
	aliases := make(map[string]string, len(repositories))
	unavailable := make(
		map[string]repository.DiscoveryEnrichment,
		len(repositories),
	)
	seen := make(map[string]struct{}, len(repositories))
	aliasIndex := 0
	for _, summary := range repositories {
		fullName := strings.ToLower(strings.TrimSpace(summary.FullName))
		if fullName == "" ||
			strings.TrimSpace(summary.Owner) == "" ||
			strings.TrimSpace(summary.Name) == "" {
			return "", nil, nil, fmt.Errorf(
				"repository enrichment identity is invalid",
			)
		}
		if _, duplicate := seen[fullName]; duplicate {
			continue
		}
		seen[fullName] = struct{}{}
		if strings.TrimSpace(summary.DefaultBranch) == "" {
			unavailable[fullName] = repository.DiscoveryEnrichment{}
			continue
		}
		alias := fmt.Sprintf("repo%d", aliasIndex)
		aliasIndex++
		aliases[alias] = fullName
		writeRepositoryEnrichmentSelection(&query, alias, summary)
	}
	query.WriteString(`  rateLimit {
    limit
    remaining
    resetAt
  }
}`)
	if query.Len() > maxRepositoryEnrichmentQueryBytes {
		return "", nil, nil, fmt.Errorf(
			"repository enrichment query exceeds %d bytes",
			maxRepositoryEnrichmentQueryBytes,
		)
	}
	return query.String(), aliases, unavailable, nil
}

func writeRepositoryEnrichmentSelection(
	query *strings.Builder,
	alias string,
	summary repository.Summary,
) {
	owner := quoteGraphQLString(summary.Owner)
	name := quoteGraphQLString(summary.Name)
	branch := summary.DefaultBranch
	expressions := map[string]string{
		"readmeMarkdown":       branch + ":README.md",
		"readmePlain":          branch + ":README",
		"readmeRST":            branch + ":README.rst",
		"readmeAsciidoc":       branch + ":README.adoc",
		"readmeJapanese":       branch + ":README.ja.md",
		"readmeJapaneseLocale": branch + ":README.ja-JP.md",
		"contributingRoot":     branch + ":CONTRIBUTING.md",
		"contributingGitHub":   branch + ":.github/CONTRIBUTING.md",
		"issueTemplates":       branch + ":.github/ISSUE_TEMPLATE",
	}
	query.WriteString("  ")
	query.WriteString(alias)
	query.WriteString(": repository(owner: ")
	query.WriteString(owner)
	query.WriteString(", name: ")
	query.WriteString(name)
	query.WriteString(`) {
    nameWithOwner
    goodFirstIssues: issues(states: OPEN, labels: ["good first issue"]) {
      totalCount
    }
    helpWantedIssues: issues(states: OPEN, labels: ["help wanted"]) {
      totalCount
    }
    codeOfConduct {
      key
    }
    securityPolicyUrl
    starterGoodFirst: issues(first: 3, states: OPEN, labels: ["good first issue"], orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes { number title url updatedAt labels(first: 10) { nodes { name } } }
    }
    starterHelpWanted: issues(first: 3, states: OPEN, labels: ["help wanted"], orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes { number title url updatedAt labels(first: 10) { nodes { name } } }
    }
    recentIssues: issues(first: 10, orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes { comments(first: 10) { nodes { authorAssociation } } }
    }
    recentMergedPullRequests: pullRequests(first: 20, states: MERGED, orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes { authorAssociation }
    }
`)
	for _, field := range []string{
		"readmeMarkdown",
		"readmePlain",
		"readmeRST",
		"readmeAsciidoc",
		"readmeJapanese",
		"readmeJapaneseLocale",
	} {
		query.WriteString("    ")
		query.WriteString(field)
		query.WriteString(": object(expression: ")
		query.WriteString(quoteGraphQLString(expressions[field]))
		query.WriteString(`) {
      ... on Blob {
        byteSize
        isBinary
        text
      }
    }
`)
	}
	for _, field := range []string{"contributingRoot", "contributingGitHub"} {
		query.WriteString("    ")
		query.WriteString(field)
		query.WriteString(": object(expression: ")
		query.WriteString(quoteGraphQLString(expressions[field]))
		query.WriteString(`) {
      ... on Blob {
        byteSize
        isBinary
      }
    }
`)
	}
	query.WriteString("    issueTemplates: object(expression: ")
	query.WriteString(quoteGraphQLString(expressions["issueTemplates"]))
	query.WriteString(`) {
      ... on Tree { entries { name type } }
    }
`)
	query.WriteString("  }\n")
}

func quoteGraphQLString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func enrichmentRateLimit(
	data map[string]json.RawMessage,
	fallback port.RateLimit,
) (port.RateLimit, error) {
	if data == nil {
		return fallback, nil
	}
	raw, exists := data["rateLimit"]
	if !exists || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fallback, nil
	}
	var graphQL graphQLRateLimit
	if err := json.Unmarshal(raw, &graphQL); err != nil {
		return port.RateLimit{}, upstreamDecodeError(
			"GitHub GraphQL repository enrichment rate limit",
			err,
		)
	}
	delete(data, "rateLimit")
	return normalizeGraphQLRateLimit(&graphQL, fallback)
}

func decodeBoundedJSON(
	body io.Reader,
	maximum int64,
	description string,
	target any,
) error {
	raw, err := io.ReadAll(io.LimitReader(body, maximum+1))
	if err != nil {
		return upstreamDecodeError(description, err)
	}
	if int64(len(raw)) > maximum {
		return upstreamDecodeError(
			description,
			fmt.Errorf("exceeds %d bytes", maximum),
		)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return upstreamDecodeError(description, err)
	}
	return nil
}

type restRepositoryDiscoveryEnvelope struct {
	TotalCount        int                           `json:"total_count"`
	IncompleteResults bool                          `json:"incomplete_results"`
	Items             []restRepositoryDiscoveryItem `json:"items"`
}

type restRepositoryDiscoveryItem struct {
	ID             int64                           `json:"id"`
	Owner          owner                           `json:"owner"`
	Name           string                          `json:"name"`
	FullName       string                          `json:"full_name"`
	Description    *string                         `json:"description"`
	HTMLURL        string                          `json:"html_url"`
	Stars          int                             `json:"stargazers_count"`
	Forks          int                             `json:"forks_count"`
	Watchers       int                             `json:"watchers_count"`
	OpenIssues     int                             `json:"open_issues_count"`
	IsFork         bool                            `json:"fork"`
	IsArchived     bool                            `json:"archived"`
	HasIssues      bool                            `json:"has_issues"`
	HasDiscussions bool                            `json:"has_discussions"`
	Language       *string                         `json:"language"`
	License        *restRepositoryDiscoveryLicense `json:"license"`
	Topics         []string                        `json:"topics"`
	DefaultBranch  string                          `json:"default_branch"`
	UpdatedAt      time.Time                       `json:"updated_at"`
	PushedAt       *time.Time                      `json:"pushed_at"`
}

type restRepositoryDiscoveryLicense struct {
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
}

func (item restRepositoryDiscoveryItem) toDomain() (
	repository.DiscoveryCandidate,
	error,
) {
	if item.ID < 0 ||
		strings.TrimSpace(item.Owner.Login) == "" ||
		strings.TrimSpace(item.Name) == "" ||
		strings.TrimSpace(item.FullName) == "" ||
		strings.TrimSpace(item.HTMLURL) == "" ||
		item.Stars < 0 ||
		item.Forks < 0 ||
		item.Watchers < 0 ||
		item.OpenIssues < 0 ||
		item.UpdatedAt.IsZero() ||
		len(item.Topics) > maxRepositoryTopics {
		return repository.DiscoveryCandidate{}, errors.New(
			"contains invalid repository fields",
		)
	}
	if err := validateAbsoluteHTTPURL(item.HTMLURL); err != nil {
		return repository.DiscoveryCandidate{}, fmt.Errorf(
			"invalid repository URL: %w",
			err,
		)
	}

	topics := make([]string, 0, len(item.Topics))
	seenTopics := make(map[string]struct{}, len(item.Topics))
	for _, rawTopic := range item.Topics {
		topic := strings.TrimSpace(rawTopic)
		if topic == "" {
			return repository.DiscoveryCandidate{}, errors.New(
				"contains a blank repository topic",
			)
		}
		key := strings.ToLower(topic)
		if _, duplicate := seenTopics[key]; duplicate {
			continue
		}
		seenTopics[key] = struct{}{}
		topics = append(topics, topic)
	}
	slices.SortFunc(topics, func(left, right string) int {
		return strings.Compare(strings.ToLower(left), strings.ToLower(right))
	})

	var pushedAt time.Time
	if item.PushedAt != nil {
		pushedAt = item.PushedAt.UTC()
	}
	var license repository.SPDXLicense
	var licenseName string
	licenseKnown := false
	if item.License != nil {
		licenseName = strings.TrimSpace(item.License.Name)
		parsed, err := repository.ParseSPDXLicense(item.License.SPDXID)
		if err == nil {
			license = parsed
			licenseKnown = true
		}
	}
	return repository.DiscoveryCandidate{
		Repository: repository.Summary{
			ID:            item.ID,
			Owner:         strings.TrimSpace(item.Owner.Login),
			Name:          strings.TrimSpace(item.Name),
			FullName:      strings.TrimSpace(item.FullName),
			Description:   stringValue(item.Description),
			URL:           item.HTMLURL,
			MainLanguage:  stringValue(item.Language),
			Stars:         item.Stars,
			Forks:         item.Forks,
			OpenIssues:    item.OpenIssues,
			IsFork:        item.IsFork,
			IsArchived:    item.IsArchived,
			DefaultBranch: strings.TrimSpace(item.DefaultBranch),
			UpdatedAt:     item.UpdatedAt.UTC(),
			PushedAt:      pushedAt,
		},
		Topics:           topics,
		License:          license,
		LicenseName:      licenseName,
		LicenseKnown:     licenseKnown,
		Watchers:         item.Watchers,
		HasIssuesEnabled: item.HasIssues,
		HasDiscussions:   item.HasDiscussions,
	}, nil
}

type graphQLRepositoryEnrichmentEnvelope struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []graphQLError             `json:"errors"`
}

type graphQLCodeOfConduct struct {
	Key string `json:"key"`
}

type graphQLRepositoryEnrichmentNode struct {
	NameWithOwner        string                          `json:"nameWithOwner"`
	GoodFirstIssues      graphQLTotalCount               `json:"goodFirstIssues"`
	HelpWantedIssues     graphQLTotalCount               `json:"helpWantedIssues"`
	CodeOfConduct        *graphQLCodeOfConduct           `json:"codeOfConduct"`
	SecurityPolicyURL    *string                         `json:"securityPolicyUrl"`
	ReadmeMarkdown       *graphQLRepositoryDiscoveryBlob `json:"readmeMarkdown"`
	ReadmePlain          *graphQLRepositoryDiscoveryBlob `json:"readmePlain"`
	ReadmeRST            *graphQLRepositoryDiscoveryBlob `json:"readmeRST"`
	ReadmeAsciidoc       *graphQLRepositoryDiscoveryBlob `json:"readmeAsciidoc"`
	ReadmeJapanese       *graphQLRepositoryDiscoveryBlob `json:"readmeJapanese"`
	ReadmeJapaneseLocale *graphQLRepositoryDiscoveryBlob `json:"readmeJapaneseLocale"`
	ContributingRoot     *graphQLRepositoryDiscoveryBlob `json:"contributingRoot"`
	ContributingGitHub   *graphQLRepositoryDiscoveryBlob `json:"contributingGitHub"`
	IssueTemplates       *graphQLRepositoryDiscoveryTree `json:"issueTemplates"`
	StarterGoodFirst     graphQLStarterIssueConnection   `json:"starterGoodFirst"`
	StarterHelpWanted    graphQLStarterIssueConnection   `json:"starterHelpWanted"`
	RecentIssues         graphQLRecentIssueConnection    `json:"recentIssues"`
	RecentMergedPRs      graphQLRecentPullConnection     `json:"recentMergedPullRequests"`
}

type graphQLRepositoryDiscoveryTree struct {
	Entries []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"entries"`
}

type graphQLStarterIssueConnection struct {
	Nodes []graphQLStarterIssue `json:"nodes"`
}

type graphQLStarterIssue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updatedAt"`
	Labels    struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
}

type graphQLRecentIssueConnection struct {
	Nodes []struct {
		Comments struct {
			Nodes []struct {
				AuthorAssociation string `json:"authorAssociation"`
			} `json:"nodes"`
		} `json:"comments"`
	} `json:"nodes"`
}

type graphQLRecentPullConnection struct {
	Nodes []struct {
		AuthorAssociation string `json:"authorAssociation"`
	} `json:"nodes"`
}

type graphQLRepositoryDiscoveryBlob struct {
	ByteSize int     `json:"byteSize"`
	IsBinary bool    `json:"isBinary"`
	Text     *string `json:"text"`
}

func (node graphQLRepositoryEnrichmentNode) toDomain(
	expectedFullName string,
) (repository.DiscoveryEnrichment, error) {
	if !strings.EqualFold(
		strings.TrimSpace(node.NameWithOwner),
		expectedFullName,
	) {
		return repository.DiscoveryEnrichment{}, errors.New(
			"repository enrichment identity does not match",
		)
	}
	if node.GoodFirstIssues.TotalCount < 0 ||
		node.HelpWantedIssues.TotalCount < 0 {
		return repository.DiscoveryEnrichment{}, errors.New(
			"repository enrichment contains invalid issue counts",
		)
	}
	readmes := []*graphQLRepositoryDiscoveryBlob{
		node.ReadmeJapaneseLocale,
		node.ReadmeJapanese,
		node.ReadmeMarkdown,
		node.ReadmePlain,
		node.ReadmeRST,
		node.ReadmeAsciidoc,
	}
	available := false
	contentAvailable := false
	sampled := false
	var content strings.Builder
	for _, blob := range readmes {
		if blob == nil {
			continue
		}
		if blob.ByteSize < 0 {
			return repository.DiscoveryEnrichment{}, errors.New(
				"repository enrichment contains a negative blob size",
			)
		}
		available = true
		if blob.IsBinary || blob.Text == nil {
			continue
		}
		contentAvailable = true
		if content.Len() > 0 {
			content.WriteByte('\n')
		}
		content.WriteString(*blob.Text)
		if blob.ByteSize > len(*blob.Text) {
			sampled = true
		}
	}
	return repository.DiscoveryEnrichment{
		Available:              true,
		READMEAvailable:        available,
		READMEContentAvailable: contentAvailable,
		READMEText:             content.String(),
		READMEContentSampled:   sampled,
		ContributingAvailable: node.ContributingRoot != nil ||
			node.ContributingGitHub != nil,
		GoodFirstIssues:  node.GoodFirstIssues.TotalCount,
		HelpWantedIssues: node.HelpWantedIssues.TotalCount,
		HasCodeOfConduct: node.CodeOfConduct != nil,
		HasSecurityPolicy: node.SecurityPolicyURL != nil &&
			strings.TrimSpace(*node.SecurityPolicyURL) != "",
		HasIssueTemplate:      hasIssueTemplate(node.IssueTemplates),
		HasTestInstructions:   hasTestInstructions(content.String()),
		HasMaintainerResponse: hasMaintainerResponse(node.RecentIssues),
		HasExternalMergedPR:   hasExternalMergedPullRequest(node.RecentMergedPRs),
		StarterIssues:         normalizeStarterIssues(node.StarterGoodFirst, node.StarterHelpWanted),
	}, nil
}

func hasIssueTemplate(tree *graphQLRepositoryDiscoveryTree) bool {
	if tree == nil {
		return false
	}
	for _, entry := range tree.Entries {
		name := strings.ToLower(strings.TrimSpace(entry.Name))
		if name == "config.yml" || name == "config.yaml" || strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			return true
		}
	}
	return false
}

func hasTestInstructions(readme string) bool {
	text := strings.ToLower(readme)
	for _, command := range []string{"npm test", "pnpm test", "yarn test", "go test", "pytest", "cargo test", "mvn test", "gradle test", "bundle exec rspec"} {
		if strings.Contains(text, command) {
			return true
		}
	}
	return strings.Contains(text, "## test") || strings.Contains(text, "## running tests")
}

func hasMaintainerResponse(connection graphQLRecentIssueConnection) bool {
	for _, issueNode := range connection.Nodes {
		for _, comment := range issueNode.Comments.Nodes {
			switch strings.ToUpper(comment.AuthorAssociation) {
			case "OWNER", "MEMBER", "COLLABORATOR":
				return true
			}
		}
	}
	return false
}

func hasExternalMergedPullRequest(connection graphQLRecentPullConnection) bool {
	for _, pull := range connection.Nodes {
		switch strings.ToUpper(pull.AuthorAssociation) {
		case "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR":
			return true
		}
	}
	return false
}

func normalizeStarterIssues(connections ...graphQLStarterIssueConnection) []repository.StarterIssue {
	result := make([]repository.StarterIssue, 0, 3)
	seen := make(map[int]struct{}, 6)
	for _, connection := range connections {
		for _, item := range connection.Nodes {
			if len(result) == 3 {
				return result
			}
			if item.Number < 1 || strings.TrimSpace(item.Title) == "" || item.UpdatedAt.IsZero() || validateAbsoluteHTTPURL(item.URL) != nil {
				continue
			}
			if _, exists := seen[item.Number]; exists {
				continue
			}
			seen[item.Number] = struct{}{}
			labels := make([]string, 0, len(item.Labels.Nodes))
			for _, label := range item.Labels.Nodes {
				if value := strings.TrimSpace(label.Name); value != "" {
					labels = append(labels, value)
				}
			}
			result = append(result, repository.StarterIssue{Number: item.Number, Title: strings.TrimSpace(item.Title), URL: item.URL, UpdatedAt: item.UpdatedAt.UTC(), Labels: labels})
		}
	}
	return result
}
