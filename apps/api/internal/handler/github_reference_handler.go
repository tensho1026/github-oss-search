package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
	"github.com/tensho1026/github-issue-search/apps/api/internal/transport/response"
	"github.com/tensho1026/github-issue-search/apps/api/internal/usecase"
)

const maximumReferenceObservationBytes = 1024

// GitHubReferenceHandler presents one public, read-only state observation.
type GitHubReferenceHandler struct {
	observe   usecase.ObserveGitHubReference
	responder response.Responder
}

// NewGitHubReferenceHandler binds the observer to HTTP transport.
func NewGitHubReferenceHandler(
	observe usecase.ObserveGitHubReference,
	responder response.Responder,
) GitHubReferenceHandler {
	return GitHubReferenceHandler{observe: observe, responder: responder}
}

type githubReferenceRequest struct {
	Kind           port.GitHubReferenceKind `json:"kind"`
	Owner          string                   `json:"owner"`
	RepositoryName string                   `json:"repositoryName"`
	Number         int                      `json:"number"`
}

// Observe accepts a bounded JSON reference and never mutates GitHub or stores
// the observation.
func (handler GitHubReferenceHandler) Observe(ctx *gin.Context) {
	request, err := decodeStrictJSONBody[githubReferenceRequest](ctx, strictJSONOptions{
		description:      "GitHub reference observation request",
		maximumBytes:     maximumReferenceObservationBytes,
		rejectNullFields: true,
	})
	if err != nil {
		handler.invalidRequest(ctx, err)
		return
	}
	result, err := handler.observe.Execute(ctx.Request.Context(), usecase.ObserveGitHubReferenceInput{
		Kind: request.Kind, Owner: request.Owner,
		RepositoryName: request.RepositoryName, Number: request.Number,
	})
	if err != nil {
		handler.responder.Error(ctx, err)
		return
	}
	var remaining *int
	if result.RateLimit.Known {
		remaining = &result.RateLimit.Remaining
	}
	handler.responder.DataWithMeta(ctx, http.StatusOK, struct {
		State port.GitHubReferenceState `json:"state"`
	}{State: result.State}, response.MetaOptions{RateLimitRemaining: remaining})
}

func (handler GitHubReferenceHandler) invalidRequest(ctx *gin.Context, err error) {
	handler.responder.Error(ctx, apperror.Wrap(
		apperror.CodeInvalidRequest,
		"GitHub reference observation request is invalid",
		http.StatusBadRequest,
		err,
	))
}
