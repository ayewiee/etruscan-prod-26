package pagination

type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func NewPagination(pageInp, sizeInp *int) Pagination {
	page := 0
	size := 20

	if pageInp != nil {
		page = *pageInp
	}
	if sizeInp != nil {
		size = *sizeInp
	}

	return Pagination{Page: page, Size: size}
}

func (p Pagination) Limit() int {
	return p.Size
}

func (p Pagination) Offset() int {
	return p.Page * p.Size
}
