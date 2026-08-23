package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/issue"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
)

const operationObserveReference = "reference.observe"

type referenceStateResponse struct {
	State    string  `json:"state"`
	MergedAt *string `json:"merged_at"`
}

// ObserveReference checks only public existence and state through one REST
// request. A GitHub 404 is deliberately reported as inaccessible because the
// anonymous public boundary cannot distinguish deleted from private objects.
func (c *Client) ObserveReference(
	ctx context.Context,
	kind port.GitHubReferenceKind,
	owner string,
	repositoryName string,
	number int,
) (port.GitHubReferenceObservation, error) {
	validationNumber := number
	if kind == port.GitHubReferenceRepository {
		validationNumber = 1
	}
	reference, err := issue.NewReference(owner, repositoryName, validationNumber)
	if err != nil {
		return port.GitHubReferenceObservation{}, err
	}
	endpoint := *c.baseURL
	endpoint.Path = path.Join(
		endpoint.Path,
		"repos",
		url.PathEscape(reference.Owner()),
		url.PathEscape(reference.RepositoryName()),
	)
	switch kind {
	case port.GitHubReferenceRepository:
	case port.GitHubReferenceIssue:
		endpoint.Path = path.Join(endpoint.Path, "issues", strconv.Itoa(number))
	case port.GitHubReferencePullRequest:
		endpoint.Path = path.Join(endpoint.Path, "pulls", strconv.Itoa(number))
	default:
		return port.GitHubReferenceObservation{}, fmt.Errorf("unsupported GitHub reference kind")
	}
	response, err := c.do(ctx, operationObserveReference, endpoint.String())
	if err != nil {
		return port.GitHubReferenceObservation{}, err
	}
	defer response.Body.Close()
	rateLimit := parseRateLimit(response.Header)
	if response.StatusCode == http.StatusNotFound {
		return port.GitHubReferenceObservation{
			State: port.GitHubReferenceInaccessible, RateLimit: rateLimit,
		}, nil
	}
	if statusErr := responseError(response.StatusCode, rateLimit); statusErr != nil {
		return port.GitHubReferenceObservation{}, statusErr
	}
	if kind == port.GitHubReferenceRepository {
		return port.GitHubReferenceObservation{
			State: port.GitHubReferenceAvailable, RateLimit: rateLimit,
		}, nil
	}
	var payload referenceStateResponse
	if decodeErr := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(&payload); decodeErr != nil {
		return port.GitHubReferenceObservation{}, upstreamDecodeError("GitHub reference response", decodeErr)
	}
	state := port.GitHubReferenceClosed
	if payload.State == "open" {
		state = port.GitHubReferenceOpen
	} else if kind == port.GitHubReferencePullRequest && payload.MergedAt != nil {
		state = port.GitHubReferenceMerged
	}
	return port.GitHubReferenceObservation{State: state, RateLimit: rateLimit}, nil
}
