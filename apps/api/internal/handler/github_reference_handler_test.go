package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/port"
	"github.com/tensho1026/github-issue-search/apps/api/internal/transport/response"
	"github.com/tensho1026/github-issue-search/apps/api/internal/usecase"
)

type observeReferenceStub struct {
	input  usecase.ObserveGitHubReferenceInput
	result port.GitHubReferenceObservation
	err    error
}

func (stub *observeReferenceStub) Execute(
	_ context.Context,
	input usecase.ObserveGitHubReferenceInput,
) (port.GitHubReferenceObservation, error) {
	stub.input = input
	return stub.result, stub.err
}

func TestGitHubReferenceHandlerObservesAndRejectsInvalidBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &observeReferenceStub{result: port.GitHubReferenceObservation{
		State:     port.GitHubReferenceOpen,
		RateLimit: port.RateLimit{Known: true, Remaining: 42},
	}}
	handler := NewGitHubReferenceHandler(stub, response.NewResponder())
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"kind":"issue","owner":"acme","repositoryName":"rocket","number":7}`,
	))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.Observe(ctx)
	if recorder.Code != http.StatusOK ||
		!strings.Contains(recorder.Body.String(), `"state":"open"`) ||
		!strings.Contains(recorder.Body.String(), `"rateLimitRemaining":42`) ||
		stub.input.Number != 7 {
		t.Fatalf("response = %d %s; input = %+v", recorder.Code, recorder.Body.String(), stub.input)
	}

	invalidRecorder := httptest.NewRecorder()
	invalidContext, _ := gin.CreateTestContext(invalidRecorder)
	invalidContext.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"unknown":true}`))
	invalidContext.Request.Header.Set("Content-Type", "application/json")
	handler.Observe(invalidContext)
	if invalidRecorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid status = %d", invalidRecorder.Code)
	}

	stub.err = apperror.Wrap(apperror.CodeGitHubAPI, "failed", http.StatusBadGateway, errors.New("upstream"))
	errorRecorder := httptest.NewRecorder()
	errorContext, _ := gin.CreateTestContext(errorRecorder)
	errorContext.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"kind":"repository","owner":"acme","repositoryName":"rocket","number":0}`,
	))
	errorContext.Request.Header.Set("Content-Type", "application/json")
	handler.Observe(errorContext)
	if errorRecorder.Code != http.StatusBadGateway {
		t.Fatalf("error status = %d", errorRecorder.Code)
	}
}
