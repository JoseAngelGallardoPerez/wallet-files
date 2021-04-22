package http

import (
	"github.com/Confialink/wallet-files/internal/database"
	"github.com/Confialink/wallet-pkg-list_params"
)

var showListOutputFields = []interface{}{
	"ID",
	"CreatedAt",
	"UpdatedAt",
	"UserId",
	"Path",
	"Filename",
	"Bucket",
	"Storage",
	"ContentType",
	"Size",
	"IsAdminOnly",
	"IsPrivate",
}

func getListParams(query string) *list_params.ListParams {
	params := list_params.NewListParamsFromQuery(query, database.FileModel{})
	params.AllowSelectFields(showListOutputFields)
	params.Pagination.PageSize = 0
	addIncludes(params)
	addSortings(params)
	allowFilters(params)
	addFilters(params)
	return params
}

func addIncludes(params *list_params.ListParams) {}

func addSortings(params *list_params.ListParams) {
	params.AllowSortings([]string{"createdAt"})
}

func allowFilters(params *list_params.ListParams) {}

func addFilters(params *list_params.ListParams) {}
