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
	if p.Page < 1 || p.Limit < 1 {
		p.SetToDefault()
	}
	if p.Limit > MaximumLimit {
		p.Limit = MaximumLimit
	}
}

// SetTotalPage will set TotalPage value
func (p *Pagination) SetTotalPage() {
	p.ValidatePagination()
	if p.TotalData > 0 && p.TotalData < p.Limit {
		p.Limit = p.TotalData
	}

	p.TotalPage = int(math.Ceil(float64(p.TotalData) / float64(p.Limit)))
}
