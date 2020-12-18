package pagination

import "math"

type Paginator interface {
	Paginate()
}

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaximumLimit = 100
)

// Paginate params
func (p *Pagination) Paginate() {
	p.ValidatePagination()
	p.Offset = p.Limit * (p.Page - 1)
}

// SetToDefault will set to it's default value
func (p *Pagination) SetToDefault() {
	p.Page, p.Limit = DefaultPage, DefaultLimit
}

// ValidatePagination will validate pagination's value
func (p *Pagination) ValidatePagination() {
	if p.Page < 1 || p.Limit < 1 || p.Limit > MaximumLimit {
		p.SetToDefault()
	}
}

// SetTotalPage will set TotalPage value
func (p *Pagination) SetTotalPage() {
	if p.Total > 0 && p.Total < p.Limit {
		p.Limit = p.Total
	}

	p.TotalPage = int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}
