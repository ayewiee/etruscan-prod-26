package pagination

type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func (p Pagination) Limit() int {
	return p.Size
}

func (p Pagination) Offset() int {
	return p.Page * p.Size
}
