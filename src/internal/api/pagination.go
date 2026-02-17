package api

import (
	"etruscan/internal/api/apierrors"
	"etruscan/internal/common/pagination"
	"strconv"

	"github.com/labstack/echo/v4"
)

func ParsePagination(c echo.Context) (*pagination.Pagination, error) {
	var err error
	page := 0
	size := 20

	rawPage := c.QueryParam("page")
	if rawPage != "" {
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil || page < 0 {
			return nil, apierrors.DumbValidationError(
				"page",
				c.QueryParam("page"),
				"page must be > 0",
				err,
			)
		}

	}
	rawSize := c.QueryParam("size")
	if rawSize != "" {
		size, err = strconv.Atoi(c.QueryParam("size"))
		if err != nil || size < 1 || size > 100 {
			return nil, apierrors.DumbValidationError(
				"size",
				c.QueryParam("size"),
				"page must be between 1 and 100",
				err,
			)
		}
	}

	return &pagination.Pagination{Page: page, Size: size}, nil
}
