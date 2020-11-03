package pagination

// Pagination scheme
type Pagination struct {
	Limit  int `form:"limit" cache:"optional"`
	Page   int `form:"page" cache:"optional"`
	Offset int
}

// Paginate params
func (pagination *Pagination) Paginate() {
	if pagination.Limit == 0 || pagination.Limit < 1 {
		pagination.Limit = 10
	}
	if pagination.Page == 0 || pagination.Page < 1 {
		pagination.Page = 1
	}

	pagination.Offset = (pagination.Page - 1) * pagination.Limit
}
