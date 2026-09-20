package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
	"github.com/tensho1026/github-issue-search/apps/api/internal/usecase"
)

// ListProfileSnapshots returns the authenticated account's bounded monthly history.
func (handler AccountHandler) ListProfileSnapshots(ctx *gin.Context) {
	accountID, ok := handler.accountID(ctx)
	if !ok {
		return
	}
	snapshots, err := handler.workspace.ListProfileSnapshots(ctx.Request.Context(), accountID)
	if err != nil {
		handler.responder.Error(ctx, err)
		return
	}
	handler.responder.Data(ctx, http.StatusOK, struct {
		Items []profileSnapshotResponse `json:"items"`
	}{Items: profileSnapshotResponses(snapshots)})
}

// UpsertProfileSnapshot stores or replaces the current UTC calendar month.
func (handler AccountHandler) UpsertProfileSnapshot(ctx *gin.Context) {
	accountID, ok := handler.accountID(ctx)
	if !ok {
		return
	}
	request, err := decodeAccountBody[profileSnapshotWriteRequest](ctx)
	if err != nil {
		handler.invalidRequest(ctx, err)
		return
	}
	proficiency := make([]account.SnapshotProficiency, len(request.Proficiency))
	for index, value := range request.Proficiency {
		proficiency[index] = account.SnapshotProficiency{Name: value.Name, Level: value.Level}
	}
	snapshot, err := handler.workspace.UpsertProfileSnapshot(ctx.Request.Context(), accountID, usecase.ProfileSnapshotInput{
		Languages: request.Languages, Frameworks: request.Frameworks,
		OSSActivity: request.OSSActivity, MergedPullRequests: request.MergedPullRequests,
		Proficiency: proficiency, CompletedQuests: request.CompletedQuests,
		CurrentStreak: request.CurrentStreak, LongestStreak: request.LongestStreak,
	})
	if err != nil {
		handler.responder.Error(ctx, err)
		return
	}
	handler.responder.Data(ctx, http.StatusOK, newProfileSnapshotResponse(snapshot))
}

type profileSnapshotProficiencyResponse struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type profileSnapshotResponse struct {
	Month              time.Time                            `json:"month"`
	Languages          []string                             `json:"languages"`
	Frameworks         []string                             `json:"frameworks"`
	OSSActivity        int                                  `json:"ossActivity"`
	MergedPullRequests int                                  `json:"mergedPullRequests"`
	Proficiency        []profileSnapshotProficiencyResponse `json:"proficiency"`
	CompletedQuests    int                                  `json:"completedQuests"`
	CurrentStreak      int                                  `json:"currentStreak"`
	LongestStreak      int                                  `json:"longestStreak"`
	CreatedAt          time.Time                            `json:"createdAt"`
	UpdatedAt          time.Time                            `json:"updatedAt"`
}

type profileSnapshotWriteRequest struct {
	Languages          []string                             `json:"languages"`
	Frameworks         []string                             `json:"frameworks"`
	OSSActivity        int                                  `json:"ossActivity"`
	MergedPullRequests int                                  `json:"mergedPullRequests"`
	Proficiency        []profileSnapshotProficiencyResponse `json:"proficiency"`
	CompletedQuests    int                                  `json:"completedQuests"`
	CurrentStreak      int                                  `json:"currentStreak"`
	LongestStreak      int                                  `json:"longestStreak"`
}

func newProfileSnapshotResponse(snapshot account.ProfileSnapshot) profileSnapshotResponse {
	proficiency := make([]profileSnapshotProficiencyResponse, len(snapshot.Proficiency))
	for index, value := range snapshot.Proficiency {
		proficiency[index] = profileSnapshotProficiencyResponse{Name: value.Name, Level: value.Level}
	}
	return profileSnapshotResponse{Month: snapshot.Month, Languages: append([]string(nil), snapshot.Languages...), Frameworks: append([]string(nil), snapshot.Frameworks...), OSSActivity: snapshot.OSSActivity, MergedPullRequests: snapshot.MergedPullRequests, Proficiency: proficiency, CompletedQuests: snapshot.CompletedQuests, CurrentStreak: snapshot.CurrentStreak, LongestStreak: snapshot.LongestStreak, CreatedAt: snapshot.CreatedAt, UpdatedAt: snapshot.UpdatedAt}
}

func profileSnapshotResponses(snapshots []account.ProfileSnapshot) []profileSnapshotResponse {
	result := make([]profileSnapshotResponse, len(snapshots))
	for index, snapshot := range snapshots {
		result[index] = newProfileSnapshotResponse(snapshot)
	}
	return result
}
