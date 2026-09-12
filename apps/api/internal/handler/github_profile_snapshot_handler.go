package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/user"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
	"github.com/tensho1026/github-issue-search/apps/api/internal/transport/response"
	"github.com/tensho1026/github-issue-search/apps/api/internal/usecase"
)

// GitHubProfileSnapshotHandler serves the combined profile page payload.
type GitHubProfileSnapshotHandler struct {
	snapshot  usecase.GetGitHubProfileSnapshot
	responder response.Responder
}

// NewGitHubProfileSnapshotHandler binds the aggregate use case to a responder.
func NewGitHubProfileSnapshotHandler(
	snapshot usecase.GetGitHubProfileSnapshot,
	responder response.Responder,
) GitHubProfileSnapshotHandler {
	return GitHubProfileSnapshotHandler{
		snapshot:  snapshot,
		responder: responder,
	}
}

// Get handles one cancellable public profile snapshot request.
func (h GitHubProfileSnapshotHandler) Get(ctx *gin.Context) {
	username, err := user.ParseUsername(ctx.Param("username"))
	if err != nil {
		h.responder.Error(ctx, apperror.Wrap(
			apperror.CodeInvalidRequest,
			"GitHub username is invalid",
			http.StatusBadRequest,
			err,
		))
		return
	}

	output, err := h.snapshot.Execute(ctx.Request.Context(), username)
	if err != nil {
		h.responder.Error(ctx, err)
		return
	}

	remaining := profileSnapshotRateLimitRemaining(
		output.User.RateLimit,
		output.Analysis.RateLimit,
	)
	h.responder.DataWithMeta(
		ctx,
		http.StatusOK,
		githubProfileSnapshotResponse{
			User:     newGitHubUserResponse(output.User.Profile, output.User.Repositories),
			Analysis: newGitHubProfileAnalysisResponse(output.Analysis.Analysis),
		},
		response.MetaOptions{RateLimitRemaining: remaining},
	)
}

type githubProfileSnapshotResponse struct {
	User     githubUserResponse            `json:"user"`
	Analysis githubProfileAnalysisResponse `json:"analysis"`
}

func profileSnapshotRateLimitRemaining(
	left port.RateLimit,
	right port.RateLimit,
) *int {
	if !left.Known && !right.Known {
		return nil
	}
	if !left.Known || (right.Known && right.Remaining < left.Remaining) {
		return &right.Remaining
	}
	return &left.Remaining
}
