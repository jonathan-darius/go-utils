package pagination

// Pagination type schema
type Pagination struct {
	Limit  int `form:"limit" cache:"optional"`
	Page   int `form:"page" cache:"optional"`
	Offset int
}
type Paginator interface {
	Paginate()
}

// Paginate params
func (p *Pagination) Paginate() {
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Page < 1 {
		p.Page = 1
	}

	p.Offset = p.Limit * (p.Page - 1)
}
