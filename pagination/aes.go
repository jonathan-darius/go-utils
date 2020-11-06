package pagination

type Pagination struct {
	Limit  int `form:"limit"`
	Page   int `form:"page"`
	Offset int
}

type Paginator interface {
	Paginate()
}

// Pagination params
func (p *Pagination) Paginate() {
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Page < 1 {
		p.Page = 1
	}
	p.Offset = p.Limit * (p.Page - 1)
}
