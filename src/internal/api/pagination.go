package api

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/common/pagination"
	"strconv"

	"github.com/labstack/echo/v4"
)

func ParsePagination(c echo.Context) (pagination.Pagination, error) {
	var page, size *int

	rawPage := c.QueryParam("page")
	if rawPage != "" {
		parsedPage, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || parsedPage < 0 {
			return pagination.Pagination{}, apierrors.DumbValidationError(
				"page",
				c.QueryParam("page"),
				"page must be > 0",
				err,
			)
		}
		page = &parsedPage

	}
	rawSize := c.QueryParam("size")
	if rawSize != "" {
		parsedSize, err := strconv.Atoi(c.QueryParam("size"))
		if err != nil || parsedSize < 1 || parsedSize > 100 {
			return pagination.Pagination{}, apierrors.DumbValidationError(
				"size",
				c.QueryParam("size"),
				"page must be between 1 and 100",
				err,
			)
		}
		size = &parsedSize
	}

	return pagination.NewPagination(page, size), nil
}
