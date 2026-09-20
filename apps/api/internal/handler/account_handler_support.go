package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tensho1026/github-issue-search/apps/api/internal/domain/account"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/apperror"
	"github.com/tensho1026/github-issue-search/apps/api/internal/platform/requestcontext"
)

func (handler AccountHandler) accountID(
	ctx *gin.Context,
) (account.ID, bool) {
	if handler.workspace == nil {
		handler.responder.Error(ctx, apperror.New(
			apperror.CodeAuthUnavailable,
			"Account features are not configured",
			http.StatusServiceUnavailable,
		))
		return account.ID{}, false
	}
	principal, ok := requestcontext.Principal(ctx.Request.Context())
	if !ok {
		handler.responder.Error(ctx, apperror.New(
			apperror.CodeAuthentication,
			"Authentication is required",
			http.StatusUnauthorized,
		))
		return account.ID{}, false
	}
	return principal.Session.AccountID, true
}

func (handler AccountHandler) ownedMutationTarget(
	ctx *gin.Context,
) (account.ID, account.ResourceID, int64, bool) {
	accountID, ok := handler.accountID(ctx)
	if !ok {
		return account.ID{}, account.ResourceID{}, 0, false
	}
	rawID := ctx.Param("bookmarkID")
	if rawID == "" {
		rawID = ctx.Param("savedSearchID")
	}
	if rawID == "" {
		rawID = ctx.Param("issueClaimID")
	}
	resourceID, err := account.ParseResourceID(rawID)
	if err != nil {
		handler.invalidRequest(ctx, err)
		return account.ID{}, account.ResourceID{}, 0, false
	}
	version, err := parseRequiredVersion(ctx)
	if err != nil {
		handler.invalidRequest(ctx, err)
		return account.ID{}, account.ResourceID{}, 0, false
	}
	return accountID, resourceID, version, true
}

func (handler AccountHandler) invalidRequest(
	ctx *gin.Context,
	err error,
) {
	if err == nil {
		err = account.ErrInvalidFeatureInput
	}
	handler.responder.Error(ctx, apperror.Wrap(
		apperror.CodeInvalidRequest,
		"Account feature request is invalid",
		http.StatusBadRequest,
		err,
	))
}

func decodeAccountBody[T any](ctx *gin.Context) (T, error) {
	return decodeStrictJSONBody[T](ctx, strictJSONOptions{
		description:  "account request",
		maximumBytes: maximumAccountRequestBytes,
	})
}

func parseAccountPage(ctx *gin.Context) (account.Page, error) {
	page, perPage, err := parsePaginationQuery(
		ctx,
		1,
		account.DefaultPageSize,
	)
	if err != nil {
		return account.Page{}, err
	}
	return account.NewPage(page, perPage)
}

func parseIssueClaimListQuery(
	ctx *gin.Context,
) (account.Page, account.IssueClaimFilter, error) {
	query := ctx.Request.URL.Query()
	for key := range query {
		if key != "page" && key != "perPage" && key != "filter" {
			return account.Page{}, "", fmt.Errorf(
				"unsupported query parameter %q",
				key,
			)
		}
	}
	pageNumber, err := parseSingleQueryInteger(query["page"], 1)
	if err != nil {
		return account.Page{}, "", fmt.Errorf("page: %w", err)
	}
	perPage, err := parseSingleQueryInteger(
		query["perPage"],
		account.DefaultPageSize,
	)
	if err != nil {
		return account.Page{}, "", fmt.Errorf("perPage: %w", err)
	}
	page, err := account.NewPage(pageNumber, perPage)
	if err != nil {
		return account.Page{}, "", err
	}
	filterValue := "all"
	if values, exists := query["filter"]; exists {
		if len(values) != 1 || values[0] == "" {
			return account.Page{}, "", fmt.Errorf(
				"filter must be provided exactly once",
			)
		}
		filterValue = values[0]
	}
	filter, err := account.NewIssueClaimFilter(filterValue)
	if err != nil {
		return account.Page{}, "", err
	}
	return page, filter, nil
}

func parseRequiredVersion(ctx *gin.Context) (int64, error) {
	query := ctx.Request.URL.Query()
	for key := range query {
		if key != "version" {
			return 0, fmt.Errorf("unsupported query parameter %q", key)
		}
	}
	values := query["version"]
	if len(values) != 1 || values[0] == "" {
		return 0, fmt.Errorf("version must be provided exactly once")
	}
	version, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || version < 1 {
		return 0, fmt.Errorf("version must be a positive integer")
	}
	return version, nil
}
